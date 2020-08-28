//
//
//

package log

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/ondi/go-queue"
)

type Convert interface {
	Convert(buf *bytes.Buffer, level string, format string, args ...interface{}) (n int, err error)
}

type Http_t struct {
	q       queue.Queue
	pool    sync.Pool
	convert Convert
	url     string
	header  http.Header
	client  *http.Client
}

// this is working example for Convert interface
type Message_t struct {
	ApplicationName string          `json:"ApplicationName"`
	Environment     string          `json:"Environment"`
	Level           string          `json:"Level"`
	Data            json.RawMessage `json:"Data,omitempty"`
	Message         json.RawMessage `json:"Message,omitempty"`

	// if CallDepth > 0 Location = "file:line" from runtime.Caller(CallDepth)
	CallDepth int    `json:"-"`
	Location  string `json:"location,omitempty"`
}

// self is copy
func (self Message_t) Convert(buf *bytes.Buffer, level string, format string, args ...interface{}) (n int, err error) {
	self.Level = level
	if len(format) == 0 {
		if self.Data, err = json.Marshal(args); err != nil {
			return
		}
	} else {
		if self.Message, err = json.Marshal(fmt.Sprintf(format, args...)); err != nil {
			return
		}
	}
	if self.CallDepth > 0 {
		if _, file, line, ok := runtime.Caller(self.CallDepth); ok {
			self.Location = fmt.Sprintf("%s:%d", path.Base(file), line)
		}
	}
	err = json.NewEncoder(buf).Encode(self)
	n = buf.Len()
	return
}

func DefaultTransport(timeout time.Duration) http.RoundTripper {
	return &http.Transport{
		DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
		TLSHandshakeTimeout:   timeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
}

func NewHttp(tr http.RoundTripper, queue_size int, writers int, post_url string, convert Convert, timeout time.Duration, header http.Header) (self *Http_t) {
	self = &Http_t{}
	self.q = queue.New(queue_size)
	self.pool = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	self.convert = convert
	self.url = post_url
	self.header = header.Clone()
	self.client = &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	for i := 0; i < writers; i++ {
		go self.writer()
	}
	return
}

func (self *Http_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	buf := self.pool.Get().(*bytes.Buffer)
	buf.Reset()
	if n, err = self.convert.Convert(buf, level, format, args...); err != nil {
		self.pool.Put(buf)
		return
	}
	if self.q.PushBackNoWait(buf) != 0 {
		self.pool.Put(buf)
		return 0, fmt.Errorf("LOG QUEUE OVERFLOW")
	}
	return
}

func (self *Http_t) Size() int {
	return self.q.Size()
}

func (self *Http_t) writer() {
	for {
		buf, ok := self.q.PopFront()
		if ok == -1 {
			return
		}
		req, err := http.NewRequest("POST", self.url, buf.(*bytes.Buffer))
		if err != nil {
			self.pool.Put(buf)
			fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
			continue
		}
		req.Header = self.header
		resp, err := self.client.Do(req)
		self.pool.Put(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
			continue
		}
		if resp.StatusCode >= 400 {
			temp, _ := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "%v ERROR: %v %s\n", time.Now().Format("2006-01-02 15:04:05"), resp.Status, temp)
		}
		resp.Body.Close()
	}
}

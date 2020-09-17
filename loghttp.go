//
//
//

package log

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/ondi/go-queue"
)

type Converter interface {
	Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error)
}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Http_t struct {
	q        queue.Queue
	pool     sync.Pool
	post_url string
	convert  Converter
	client   Client
	header   http.Header
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
	Location  string `json:"Location,omitempty"`
}

// self is copy
func (self Message_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
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
	err = json.NewEncoder(out).Encode(self)
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

func DefaultClient(tr http.RoundTripper, timeout time.Duration) Client {
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}

func NewHttp(queue_size int, writers int, post_url string, convert Converter, client Client, header http.Header) (self *Http_t) {
	self = &Http_t{
		q:        queue.New(queue_size),
		pool:     sync.Pool{New: func() interface{} { return bytes.NewBuffer(nil) }},
		post_url: post_url,
		convert:  convert,
		client:   client,
		header:   header.Clone(),
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
	var oki int
	var err error
	var temp interface{}
	var buf *bytes.Buffer
	var req *http.Request
	var resp *http.Response
	for {
		if temp, oki = self.q.PopFront(); oki == -1 {
			return
		}
		buf = temp.(*bytes.Buffer)
		if req, err = http.NewRequest(http.MethodPost, self.post_url, buf); err != nil {
			self.pool.Put(temp)
			fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
			continue
		}
		req.Header = self.header
		if resp, err = self.client.Do(req); err != nil {
			self.pool.Put(temp)
			fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
			continue
		}
		if resp.StatusCode >= 400 {
			buf.Reset()
			buf.ReadFrom(resp.Body)
			fmt.Fprintf(os.Stderr, "%v ERROR: %v %s\n", time.Now().Format("2006-01-02 15:04:05"), resp.Status, buf.Bytes())
		}
		self.pool.Put(temp)
		resp.Body.Close()
	}
}

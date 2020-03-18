//
//
//

package log

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ondi/go-queue"
)

type Convert_t func(buf *bytes.Buffer, level string, format string, args ...interface{}) (int, error)

type Http_t struct {
	q       queue.Queue
	pool    sync.Pool
	convert Convert_t
	url     string
	header  http.Header
	client  *http.Client
}

func Convert(buf *bytes.Buffer, level string, format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(buf, level+" "+format+"\n", args...)
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

func NewHttp(tr http.RoundTripper, queue_size int, workers int, post_url string, convert Convert_t, timeout time.Duration, header http.Header) (self *Http_t) {
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
	for i := 0; i < workers; i++ {
		go self.worker()
	}
	return
}

func (self *Http_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	buf := self.pool.Get().(*bytes.Buffer)
	buf.Reset()
	if n, err = self.convert(buf, level, format, args...); err != nil {
		self.pool.Put(buf)
		return
	}
	if self.q.PushBackNoWait(buf) != 0 {
		return 0, fmt.Errorf("LOG QUEUE OVERFLOW")
	}
	return
}

func (self *Http_t) Size() int {
	return self.q.Size()
}

func (self *Http_t) worker() {
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
		} else if resp.StatusCode >= 400 {
			temp, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "%v ERROR: %v %s\n", time.Now().Format("2006-01-02 15:04:05"), resp.Status, temp)
		}
	}
}

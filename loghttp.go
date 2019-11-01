//
//
//

package log

import "os"
import "fmt"
import "time"
import "sync"
import "bytes"
import "net"
import "net/http"
import "crypto/tls"

import "github.com/ondi/go-queue"

type Convert_t func(buf * bytes.Buffer, level string, format string, args ...interface{}) (int, error)

type Http_t struct {
	q queue.Queue
	pool sync.Pool
	convert Convert_t
	url string
	headers map[string]string
	client * http.Client
}

func Convert(buf * bytes.Buffer, level string, format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(buf, level + " " + format + "\n", args...)
}

func NewHttp(queue_size int, workers int, post_url string, timeout time.Duration, convert Convert_t, headers map[string]string) (self * Http_t) {
	self = &Http_t{}
	self.q = queue.New(queue_size)
	self.pool = sync.Pool {New: func() interface{} {return new(bytes.Buffer)}}
	self.convert = convert
	self.url = post_url
	self.headers = map[string]string{}
	for k, v := range headers {
		self.headers[k] = v
	}
	self.client = &http.Client {
		// Timeout: timeout,
		Transport: &http.Transport {
			Dial: (&net.Dialer{Timeout: timeout}).Dial,
			// DisableKeepAlives: false,
			// MaxIdleConns: 10,
			// MaxIdleConnsPerHost: 10,
			// IdleConnTimeout: timeout,
			// ResponseHeaderTimeout: timeout,
			// TLSHandshakeTimeout: timeout,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for i := 0; i < workers; i++ {
		go self.worker()
	}
	return
}

func (self * Http_t) Write(level string, format string, args ...interface{}) (n int, err error) {
	buf := self.pool.Get().(* bytes.Buffer)
	buf.Reset()
	if n, err = self.convert(buf, level, format, args...); err != nil {
		self.pool.Put(buf)
		fmt.Fprintf(os.Stderr, "LOG CONVERT: %v\n", err)
		return
	}
	self.q.PushBackNoWait(buf)
	return
}

func (self * Http_t) worker() {
	for {
		buf, ok := self.q.PopFront()
		if ok == -1 {
			return
		}
		req, err := http.NewRequest("POST", self.url, buf.(* bytes.Buffer))
		if err != nil {
			self.pool.Put(buf)
			fmt.Fprintf(os.Stderr, "LOG REQUEST: %v", err)
			continue
		}
		for k, v := range self.headers {
			req.Header.Set(k, v)
		}
		resp, err := self.client.Do(req)
		self.pool.Put(buf)
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "LOG POST: %v\n", err)
		} else if resp.StatusCode >= 400 {
			fmt.Fprintf(os.Stderr, "LOG STATUS: %v\n", resp.Status)
		}
	}
}

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

type Convert_t func(buf * bytes.Buffer, level string, format string, args ...interface{}) error

type Http_t struct {
	q queue.Queue
	pool sync.Pool
	convert Convert_t
	url string
	client * http.Client
}

func Convert(buf * bytes.Buffer, level string, format string, args ...interface{}) (err error) {
	_, err = fmt.Fprintf(buf, level + " " + format + "\n", args...)
	return
}

func NewHttp(post_url string, convert Convert_t, queue_size int, timeout time.Duration, workers int) (self * Http_t) {
	self = &Http_t{}
	self.q = queue.New(queue_size)
	self.pool = sync.Pool {New: func() interface{} {return new(bytes.Buffer)}}
	self.convert = convert
	self.url = post_url
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

func (self * Http_t) Write(level string, format string, args ...interface{}) (err error) {
	buf := self.pool.Get().(* bytes.Buffer)
	buf.Reset()
	if err = self.convert(buf, level, format, args...); err != nil {
		self.pool.Put(buf)
		fmt.Fprintf(os.Stderr, "%v\n", err)
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
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		resp, err := self.client.Do(req)
		self.pool.Put(buf)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		if resp.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "%v: %v\n", self.url, resp.Status)
			continue
		}
	}
}

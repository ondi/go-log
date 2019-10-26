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

type Convert_t func(* bytes.Buffer, []byte) error

type HttpLogWriter_t struct {
	q queue.Queue
	pool sync.Pool
	convert Convert_t
	url string
	client * http.Client
}

func Convert(buf * bytes.Buffer, in []byte) (err error) {
	_, err = buf.Write(in)
	return
}

func (self * HttpLogWriter_t) Write(m []byte) (int, error) {
	buf := self.pool.Get().(* bytes.Buffer)
	buf.Reset()
	if err := self.convert(buf, m); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 0, err
	}
	self.q.PushBackNoWait(buf)
	return buf.Len(), nil
}

func (self * HttpLogWriter_t) worker() {
	for {
		m, ok := self.q.PopFront()
		if ok == -1 {
			return
		}
		req, err := http.NewRequest("POST", self.url, m.(* bytes.Buffer))
		self.pool.Put(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		if _, err = self.client.Do(req); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
	}
}

func NewHttpLogWriter(post_url string, convert Convert_t, queue_size int, timeout time.Duration, workers int) (self * HttpLogWriter_t) {
	self = &HttpLogWriter_t{}
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

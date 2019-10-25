//
//
//

package log

import "os"
import "fmt"
import "time"
import "bytes"
import "net"
import "net/http"
import "crypto/tls"

import "github.com/ondi/go-queue"

type Convert_t func([]byte) (bytes.Buffer, error)

type HttpLogWriter_t struct {
	q queue.Queue
	convert Convert_t
	url string
	client * http.Client
}

func Convert(in []byte) (buf bytes.Buffer, err error) {
	_, err = buf.Write(in)
	return
}

func (self * HttpLogWriter_t) Write(m []byte) (int, error) {
	// self.q.PushBackNoWait(m)
	self.q.PushBackNoWait(append([]byte{}, m...))
	return 0, nil
}

func (self * HttpLogWriter_t) worker() {
	for {
		m, ok := self.q.PopFront()
		if ok == -1 {
			return
		}
		buf, err := self.convert(m.([]byte))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		req, err := http.NewRequest("POST", self.url, &buf)
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

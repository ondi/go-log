//
//
//

package log

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// Default
// MaxIdleConns:        100,
// MaxIdleConnsPerHost: 2,
func DefaultTransport(timeout time.Duration, MaxIdleConns int, MaxIdleConnsPerHost int) http.RoundTripper {
	return &http.Transport{
		DialContext:         (&net.Dialer{Timeout: timeout}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		// TLSHandshakeTimeout:   timeout,
		// IdleConnTimeout:       timeout,
		// ResponseHeaderTimeout: timeout,
		// ExpectContinueTimeout: timeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

type Urls interface {
	Range() []string
}

type Urls_t struct {
	mx   sync.Mutex
	urls [][]string
	i    int
}

func NewUrls(urls ...string) (self *Urls_t) {
	self = &Urls_t{}
	self.urls = make([][]string, len(urls))
	for i := 0; i < len(urls); i++ {
		for k := i; k < len(urls)+i; k++ {
			self.urls[i] = append(self.urls[i], urls[k%len(urls)])
		}
	}
	return
}

func (self *Urls_t) Range() (res []string) {
	self.mx.Lock()
	res = self.urls[self.i]
	self.i = (self.i + 1) % len(self.urls)
	self.mx.Unlock()
	return
}

type Http_t struct {
	urls       Urls
	message    Formatter
	client     Client
	rps_limit  Rps
	header     http.Header
	post_delay time.Duration
	bulk_write int
}

type HttpOption func(self *Http_t)

func PostHeader(header http.Header) HttpOption {
	return func(self *Http_t) {
		self.header = header.Clone()
	}
}

func PostDelay(delay time.Duration) HttpOption {
	return func(self *Http_t) {
		self.post_delay = delay
	}
}

func RpsLimit(rps_limit Rps) HttpOption {
	return func(self *Http_t) {
		self.rps_limit = rps_limit
	}
}

func BulkWrite(bulk_write int) HttpOption {
	return func(self *Http_t) {
		if bulk_write > 0 {
			self.bulk_write = bulk_write
		}
	}
}

func NewHttpQueue(queue_size int, writers int, urls Urls, message Formatter, client Client, opts ...HttpOption) Queue {
	self := &Http_t{
		urls:       urls,
		message:    message,
		client:     client,
		rps_limit:  NoRps_t{},
		bulk_write: 1,
	}

	for _, opt := range opts {
		opt(self)
	}

	q := NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		q.WgAdd(1)
		go self.writer(q)
	}

	return q
}

func (self *Http_t) writer(q Queue) (err error) {
	defer q.WgDone()

	var req *http.Request
	var resp *http.Response
	var body bytes.Buffer
	msg := make([]Msg_t, self.bulk_write)
	for {
		body.Reset()
		n, oki := q.ReadLog(msg)
		for i := 0; i < n; i++ {
			if self.rps_limit.Add(msg[i].Level.Ts) == false {
				q.WriteError(1)
				fmt.Fprintf(STDERR, "LOG ERROR: %v ERROR_RPS\n", msg[i].Level.Ts.Format("2006-01-01 15:04:05"))
				continue
			}
			if _, err = self.message.FormatLog(&body, msg[i]); err != nil {
				q.WriteError(1)
				fmt.Fprintf(STDERR, "LOG ERROR: %v %v\n", msg[i].Level.Ts.Format("2006-01-01 15:04:05"), err)
				continue
			}
		}
		if body.Len() > 0 {
			for _, v := range self.urls.Range() {
				if req, err = http.NewRequest(http.MethodPost, v, bytes.NewReader(body.Bytes())); err != nil {
					continue
				}
				req.Header = self.header
				if resp, err = self.client.Do(req); err != nil {
					fmt.Fprintf(STDERR, "LOG ERROR: %v count=%v, err=%v\n", time.Now().Format("2006-01-02 15:04:05"), n, err)
					continue
				}
				resp.Body.Close()
				if resp.StatusCode >= 400 {
					fmt.Fprintf(STDERR, "LOG ERROR: %v count=%v, status=%v, body=%s\n", time.Now().Format("2006-01-02 15:04:05"), n, resp.Status, body.Bytes())
					continue
				}
				break
			}
			if err != nil || resp != nil && resp.StatusCode >= 400 {
				q.WriteError(n)
			}
		}
		if oki == -1 {
			return
		}
		time.Sleep(self.post_delay)
	}
}

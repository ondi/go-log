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
	"os"
	"sync"
	"time"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
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
	wg         sync.WaitGroup
	urls       Urls
	message    Formatter
	client     Client
	rps_limit  Rps
	header     http.Header
	post_delay time.Duration
	bulk_write int
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

func NewHttp(queue_size int, writers int, urls Urls, message Formatter, client Client, opts ...HttpOption) Queue {
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
		self.wg.Add(1)
		go self.writer(q)
	}

	return q
}

func (self *Http_t) writer(q Queue) (err error) {
	defer self.wg.Done()

	var oki int
	var req *http.Request
	var resp *http.Response
	var body bytes.Buffer
	var ms []Msg_t
	for {
		body.Reset()
		ms, oki = q.ReadLog(self.bulk_write)
		if oki == -1 {
			return
		}
		if len(ms) == 0 {
			continue
		}
		if !self.rps_limit.Add(ms[0].Ts) {
			fmt.Fprintf(os.Stderr, "LOG ERROR: %v ERROR_RPS\n", ms[0].Ts.Format("2006-01-01 15:04:05"))
			continue
		}
		for _, m := range ms {
			if _, err = self.message.FormatLog(m.Ctx, &body, m.Ts, m.Level, m.Format, m.Args...); err != nil {
				fmt.Fprintf(os.Stderr, "LOG ERROR: %v %v\n", ms[0].Ts.Format("2006-01-01 15:04:05"), err)
			}
		}
		for _, v := range self.urls.Range() {
			if req, err = http.NewRequest(http.MethodPost, v, bytes.NewReader(body.Bytes())); err != nil {
				continue
			}
			req.Header = self.header
			if resp, err = self.client.Do(req); err != nil {
				fmt.Fprintf(os.Stderr, "LOG ERROR: %v %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				fmt.Fprintf(os.Stderr, "LOG ERROR: %v %v %s\n", time.Now().Format("2006-01-02 15:04:05"), resp.Status, body.Bytes())
				continue
			}
			break
		}
		time.Sleep(self.post_delay)
	}
}

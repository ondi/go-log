//
//
//

package log

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Urls interface {
	Range() []string
}

type Headers interface {
	Header(*http.Request) error
}

// Default
// MaxIdleConns:        100,
// MaxIdleConnsPerHost: 2,
func DefaultTransport(dial_timeout time.Duration, MaxIdleConns int, MaxIdleConnsPerHost int) http.RoundTripper {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: dial_timeout}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          MaxIdleConns,
		MaxIdleConnsPerHost:   MaxIdleConnsPerHost,
		TLSHandshakeTimeout:   30 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
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

type NoHeaders_t struct{}

func (NoHeaders_t) Header(*http.Request) error {
	return nil
}

type Http_t struct {
	urls       Urls
	message    Formatter
	client     Client
	rps        Rps
	headers    Headers
	post_delay time.Duration
	bulk_write int
}

type HttpOption func(self *Http_t)

func PostHeader(headers Headers) HttpOption {
	return func(self *Http_t) {
		self.headers = headers
	}
}

func PostDelay(delay time.Duration) HttpOption {
	return func(self *Http_t) {
		self.post_delay = delay
	}
}

func RpsLimit(rps_limit Rps) HttpOption {
	return func(self *Http_t) {
		self.rps = rps_limit
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
		rps:        NoRps_t{},
		headers:    NoHeaders_t{},
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

	var n int
	var ok bool
	var req *http.Request
	var resp *http.Response
	var body bytes.Buffer
	msg := make([]Msg_t, self.bulk_write)
	for {
		body.Reset()
		if n, ok = q.ReadLog(msg); !ok {
			return
		}
		if n > 0 && self.rps.Add(msg[0].Info.Ts) == false {
			q.WriteError(1)
			continue
		}
		for i := 0; i < n; i++ {
			if _, err = self.message.FormatLog(&body, msg[i]); err != nil {
				q.WriteError(1)
				body.Reset()
				continue
			}
		}
		if body.Len() == 0 {
			continue
		}
		for _, v := range self.urls.Range() {
			if req, err = http.NewRequest(http.MethodPost, v, bytes.NewReader(body.Bytes())); err != nil {
				continue
			}
			if err = self.headers.Header(req); err != nil {
				return
			}
			if resp, err = self.client.Do(req); err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				continue
			}
			break
		}
		if err != nil || StatusCode(resp) >= 400 {
			q.WriteError(n)
		}
		time.Sleep(self.post_delay)
	}
}

func StatusCode(resp *http.Response) int {
	if resp != nil {
		return resp.StatusCode
	}
	return -1
}

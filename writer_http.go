//
//
//

package log

import (
	"bytes"
	"context"
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

type PostContext interface {
	WithTimeout(context.Context) (context.Context, context.CancelFunc)
}

type PostDelayer interface {
	Delay()
}

// Default
// MaxIdleConns:        100,
// MaxIdleConnsPerHost: 2,
func DefaultTransport(dial_timeout time.Duration, MaxIdleConns int, MaxIdleConnsPerHost int) http.RoundTripper {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: dial_timeout}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
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

type Timeout_t struct {
	timeout time.Duration
}

func (self *Timeout_t) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, self.timeout)
}

func (self *Timeout_t) Delay() {
	time.Sleep(self.timeout)
}

type NoTimeout_t struct{}

func (self NoTimeout_t) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return ctx, func() {}
}

func (self NoTimeout_t) Delay() {}

type Http_t struct {
	urls       Urls
	client     Client
	rps        Rps
	headers    Headers
	post_ctx   PostContext
	post_delay PostDelayer
	message    Formatter
	bulk_write int
}

type HttpOption func(self *Http_t)

func RpsLimit(rps_limit Rps) HttpOption {
	return func(self *Http_t) {
		self.rps = rps_limit
	}
}

func PostHeader(headers Headers) HttpOption {
	return func(self *Http_t) {
		self.headers = headers
	}
}

func PostDelay(delay time.Duration) HttpOption {
	return func(self *Http_t) {
		self.post_delay = &Timeout_t{timeout: delay}
	}
}

func PostTimeout(timeout time.Duration) HttpOption {
	return func(self *Http_t) {
		self.post_ctx = &Timeout_t{timeout: timeout}
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
		post_ctx:   NoTimeout_t{},
		post_delay: NoTimeout_t{},
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

	for {
		var body bytes.Buffer
		msg, ok := q.LogRead(self.bulk_write)
		if !ok {
			return
		}
		if self.rps.Add(msg[0].Info.Ts) == false {
			q.WriteError(len(msg))
			continue
		}
		if _, err = self.message.FormatMessage(&body, msg...); err != nil {
			q.WriteError(len(msg))
			continue
		}
		var status int
		for _, v := range self.urls.Range() {
			if status, _ = self.request(v, body.Bytes()); status >= 200 && status < 300 {
				break
			}
		}
		if status == 0 || status >= 400 {
			q.WriteError(len(msg))
		}
		self.post_delay.Delay()
	}
}

func (self *Http_t) request(URL string, body []byte) (status int, err error) {
	ctx, cancel := self.post_ctx.WithTimeout(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewReader(body))
	if err != nil {
		return
	}
	if err = self.headers.Header(req); err != nil {
		return
	}
	resp, err := self.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
	status = resp.StatusCode
	return
}

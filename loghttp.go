//
//
//

package log

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ondi/go-queue"
)

var (
	ERROR_RPS      = errors.New("RPS")
	ERROR_OVERFLOW = errors.New("OVERFLOW")
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
	mx         sync.Mutex
	wg         sync.WaitGroup
	q          queue.Queue[bytes.Buffer]
	urls       Urls
	convert    Formatter
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

func NewHttp(queue_size int, writers int, urls Urls, convert Formatter, client Client, opts ...HttpOption) Writer {
	self := &Http_t{
		urls:       urls,
		convert:    convert,
		client:     client,
		rps_limit:  NoRps_t{},
		bulk_write: 1,
	}

	self.q = queue.NewOpen[bytes.Buffer](&self.mx, queue_size)

	for _, opt := range opts {
		opt(self)
	}

	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer()
	}

	return self
}

func (self *Http_t) WriteLog(ctx context.Context, ts time.Time, level string, format string, args ...any) (n int, err error) {
	if !self.rps_limit.Add(ts) {
		err = ERROR_RPS
		fmt.Fprintf(os.Stderr, "LOG ERROR: %v %v\n", ts.Format("2006-01-01 15:04:05"), err)
		return
	}

	var buf bytes.Buffer
	if n, err = self.convert.FormatLog(ctx, &buf, ts, level, format, args...); err != nil {
		return
	}

	self.mx.Lock()
	res := self.q.PushBackNoWait(buf)
	self.mx.Unlock()

	if res != 0 {
		err = ERROR_OVERFLOW
		fmt.Fprintf(os.Stderr, "LOG ERROR: %v %v\n", ts.Format("2006-01-01 15:04:05"), err)
	}
	return
}

func (self *Http_t) Size() (res int) {
	self.mx.Lock()
	res = self.q.Size()
	self.mx.Unlock()
	return
}

func (self *Http_t) writer() (err error) {
	defer self.wg.Done()

	var oki int
	var req *http.Request
	var resp *http.Response
	var buf, body bytes.Buffer
	for {
		body.Reset()
		self.mx.Lock()
		for i := 0; i < self.bulk_write; i++ {
			if buf, oki = self.q.PopFront(); oki == 0 {
				body.ReadFrom(&buf)
			} else {
				break
			}
		}
		self.mx.Unlock()
		if oki == -1 {
			return
		}
		for _, v := range self.urls.Range() {
			if req, err = http.NewRequest(http.MethodPost, v, bytes.NewReader(body.Bytes())); err != nil {
				continue
			}
			req.Header = self.header
			if resp, err = self.client.Do(req); err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				err = errors.New(resp.Status)
				continue
			}
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "LOG ERROR: %v %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		}
		time.Sleep(self.post_delay)
	}
}

func (self *Http_t) Close() error {
	self.mx.Lock()
	self.q.Close()
	self.mx.Unlock()
	self.wg.Wait()
	return nil
}

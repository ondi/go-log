//
//
//

package log

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ondi/go-queue"
)

type Converter interface {
	Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error)
}

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
	convert    Converter
	client     Client
	rps_limit  Rps
	header     http.Header
	post_delay time.Duration
	queue_size int
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
		TLSHandshakeTimeout: timeout,
		// IdleConnTimeout:       timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
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

func NewHttp(queue_size int, writers int, urls Urls, convert Converter, client Client, opts ...HttpOption) (self *Http_t) {
	self = &Http_t{
		queue_size: queue_size,
		urls:       urls,
		convert:    convert,
		client:     client,
		rps_limit:  NoRps_t{},
	}

	self.q = queue.NewOpen[bytes.Buffer](&self.mx, queue_size)

	for _, opt := range opts {
		opt(self)
	}

	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer()
	}
	return
}

func (self *Http_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	if !self.rps_limit.Add(time.Now()) {
		return 0, fmt.Errorf("RPS")
	}
	var buf bytes.Buffer
	if n, err = self.convert.Convert(&buf, level, format, args...); err != nil {
		return
	}

	self.mx.Lock()
	if self.q.Size() >= self.queue_size {
		self.q.PopFront()
	}
	res := self.q.PushBackNoWait(buf)
	self.mx.Unlock()

	if res != 0 {
		return 0, fmt.Errorf("LOG QUEUE OVERFLOW")
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

	var req *http.Request
	var resp *http.Response
	for {
		self.mx.Lock()
		buf, oki := self.q.PopFront()
		self.mx.Unlock()
		if oki == -1 {
			return
		}
		for _, v := range self.urls.Range() {
			if req, err = http.NewRequest(http.MethodPost, v, bytes.NewReader(buf.Bytes())); err != nil {
				continue
			}
			req.Header = self.header
			if resp, err = self.client.Do(req); err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode >= 400 {
				if buf.Len() > 1024 {
					err = fmt.Errorf("%s: %s", resp.Status, buf.Bytes()[:1024])
				} else {
					err = fmt.Errorf("%s: %s", resp.Status, bytes.TrimRight(buf.Bytes(), "\r\n"))
				}
				continue
			}
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v LOG ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
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

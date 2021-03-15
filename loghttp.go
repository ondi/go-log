//
//
//

package log

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
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
	q          queue.Queue
	queue_size int

	pool    sync.Pool
	urls    Urls
	convert Converter
	client  Client
	header  http.Header

	post_delay time.Duration
	rps_limit  Rps

	wg sync.WaitGroup

	ctx        context.Context
	ctx_cancel context.CancelFunc
}

// this is working example for Convert interface
type Message_t struct {
	ApplicationName string          `json:"ApplicationName"`
	Environment     string          `json:"Environment"`
	Level           string          `json:"Level"`
	Data            json.RawMessage `json:"Data,omitempty"`
	Message         json.RawMessage `json:"Message,omitempty"`

	// if CallDepth > 0 Location = "file:line" from runtime.Caller(CallDepth)
	CallDepth int    `json:"-"`
	Location  string `json:"Location,omitempty"`
}

// self is copy
func (self Message_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
	self.Level = level
	if len(format) == 0 {
		if self.Data, err = json.Marshal(args); err != nil {
			return
		}
	} else {
		if self.Message, err = json.Marshal(fmt.Sprintf(format, args...)); err != nil {
			return
		}
	}
	if self.CallDepth > 0 {
		if _, file, line, ok := runtime.Caller(self.CallDepth); ok {
			self.Location = fmt.Sprintf("%s:%d", path.Base(file), line)
		}
	}
	err = json.NewEncoder(out).Encode(self)
	return
}

func DefaultTransport(timeout time.Duration) http.RoundTripper {
	return &http.Transport{
		DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
		TLSHandshakeTimeout:   timeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
}

func DefaultClient(tr http.RoundTripper, timeout time.Duration) Client {
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
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
		pool:       sync.Pool{New: func() interface{} { return bytes.NewBuffer(nil) }},
		urls:       urls,
		convert:    convert,
		client:     client,
		rps_limit:  NoRps_t{},
	}

	self.q = queue.NewOpen(&self.mx, queue_size)

	for _, opt := range opts {
		opt(self)
	}

	self.ctx, self.ctx_cancel = context.WithCancel(context.Background())

	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer()
	}
	return
}

func (self *Http_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	if self.rps_limit.Overflow(time.Now()) {
		return 0, fmt.Errorf("RPS")
	}
	buf := self.pool.Get().(*bytes.Buffer)
	buf.Reset()
	if n, err = self.convert.Convert(buf, level, format, args...); err != nil {
		self.pool.Put(buf)
		return
	}

	self.mx.Lock()
	if self.q.Size() >= self.queue_size {
		self.q.PopFront()
	}
	res := self.q.PushBackNoWait(buf)
	self.mx.Unlock()

	if res != 0 {
		self.pool.Put(buf)
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

func (self *Http_t) writer() {
	defer self.wg.Done()

	var err error
	var req *http.Request
	var resp *http.Response
	for {
		self.mx.Lock()
		temp, oki := self.q.PopFront()
		self.mx.Unlock()
		if oki == -1 {
			return
		}
		buf := temp.(*bytes.Buffer)
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
				err = fmt.Errorf("%s: %s", resp.Status, buf.Bytes())
				continue
			}
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v ERROR: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		}
		self.pool.Put(temp)
		select {
		case <-self.ctx.Done():
		case <-time.After(self.post_delay):
		}
	}
}

func (self *Http_t) Close() error {
	self.mx.Lock()
	self.q = queue.NewClosed(self.q.Close())
	self.mx.Unlock()

	self.ctx_cancel()
	self.wg.Wait()
	return nil
}

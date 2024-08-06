//
//
//

package log

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/ondi/go-circular"
)

type WriterContext_t struct{}

func NewWriterContext() Queue {
	return &WriterContext_t{}
}

func (self *WriterContext_t) WriteLog(m LogMsg_t) (n int, err error) {
	if v := GetLogContext(m.Ctx); v != nil {
		n, err = v.WriteLog(m)
	}
	return
}

func (self *WriterContext_t) ReadLog(p []LogMsg_t) (n int, ok bool) {
	return
}

func (self *WriterContext_t) WriteError(count int) {

}

func (self *WriterContext_t) Size() (res QueueSize_t) {
	return
}

func (self *WriterContext_t) WgAdd(int) {

}

func (self *WriterContext_t) WgDone() {

}

func (self *WriterContext_t) Close() error {
	return nil
}

// &log_ctx used for ctx.Value
var log_ctx = 1

type LogContext interface {
	ContextName() string
	WriteLog(m LogMsg_t) (n int, err error)
	ContextRange(func(level int64, format string, args ...any) bool)
	ContextReset()
}

type LogContext_t struct {
	mx    sync.Mutex
	name  string
	data  *circular.List_t[LogMsg_t]
	limit int
}

func NewLogContext(name string, limit int) LogContext {
	self := &LogContext_t{
		name:  name,
		limit: limit,
		data:  circular.New[LogMsg_t](limit),
	}
	return self
}

func (self *LogContext_t) ContextName() string {
	return self.name
}

func (self *LogContext_t) WriteLog(m LogMsg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.data.Size() >= self.limit {
		self.data.PopFront()
	}
	self.data.PushBack(m)
	return
}

func (self *LogContext_t) ContextRange(f func(level int64, format string, args ...any) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.RangeFront(func(m LogMsg_t) bool {
		return f(m.Level.Level, m.Format, m.Args...)
	})

}

func (self *LogContext_t) ContextReset() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.Reset()
}

func SetLogContext(ctx context.Context, value LogContext) context.Context {
	return context.WithValue(ctx, &log_ctx, value)
}

func GetLogContext(ctx context.Context) (value LogContext) {
	value, _ = ctx.Value(&log_ctx).(LogContext)
	return
}

func GetLogContextPayload(ctx context.Context, f func(level int64, format string, args ...any) bool) {
	if v, _ := ctx.Value(&log_ctx).(LogContext); v != nil {
		v.ContextRange(f)
	}
	return
}

type log_ctx_middleware_t struct {
	Handler http.Handler
	Limit   int
}

func NewLogCtxMiddleware(next http.Handler, limit int) http.Handler {
	self := &log_ctx_middleware_t{
		Handler: next,
		Limit:   limit,
	}
	return self
}

func (self *log_ctx_middleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), &log_ctx, NewLogContext(uuid.New().String(), self.Limit)))
	self.Handler.ServeHTTP(w, r)
}

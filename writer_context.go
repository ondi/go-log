//
//
//

package log

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/ondi/go-circular"
)

// &log_ctx used for ctx.Value
var log_ctx = 1

type RangeFn_t = func(ts time.Time, file string, line int, level_name string, level_id int64, format string, args ...any) bool

type LogContext interface {
	ContextName() string
	WriteLog(m Msg_t) (n int, err error)
	ContextRange(f RangeFn_t)
	ContextReset()
}

type LogContext_t struct {
	mx    sync.Mutex
	name  string
	data  *circular.List_t[Msg_t]
	limit int
}

func NewLogContext(name string, limit int) LogContext {
	self := &LogContext_t{
		name:  name,
		limit: limit,
		data:  circular.New[Msg_t](limit),
	}
	return self
}

func (self *LogContext_t) ContextName() string {
	return self.name
}

func (self *LogContext_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.data.Size() >= self.limit {
		self.data.PopFront()
	}
	self.data.PushBack(m)
	return
}

func (self *LogContext_t) ContextRange(f RangeFn_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.RangeFront(func(m Msg_t) bool {
		return f(m.Info.Ts, m.Info.File, m.Info.Line, m.Info.LevelName, m.Info.LevelId, m.Format, m.Args...)
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

type LogContextTrim_t struct {
	level       int64
	first_words int
}

func NewLogContextTrim(level int64, first_words int) (self *LogContextTrim_t) {
	self = &LogContextTrim_t{
		level:       level,
		first_words: first_words,
	}
	return
}

func (self *LogContextTrim_t) GetPayload(ctx context.Context) (res string) {
	if v := GetLogContext(ctx); v != nil {
		v.ContextRange(func(ts time.Time, file string, line int, level_name string, level_id int64, format string, args ...any) bool {
			if level_id < self.level {
				return true
			}
			res = FirstWords(fmt.Sprintf(format, args...), self.first_words)
			return false
		})
	}
	return
}

func FirstWords(in string, count int) string {
	for i, v := range in {
		if unicode.IsSpace(v) {
			count--
			if count <= 0 {
				return in[:i]
			}
		}
	}
	return in
}

type ctx_middleware_t struct {
	Handler http.Handler
	Limit   int
}

func NewContextMiddleware(next http.Handler, limit int) http.Handler {
	self := &ctx_middleware_t{
		Handler: next,
		Limit:   limit,
	}
	return self
}

func (self *ctx_middleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), &log_ctx, NewLogContext(uuid.New().String(), self.Limit)))
	self.Handler.ServeHTTP(w, r)
}

type WriterContext_t struct {
	queue_write atomic.Int64
}

func NewWriterContext() Queue {
	return &WriterContext_t{}
}

func (self *WriterContext_t) LogWrite(m Msg_t) (n int, err error) {
	if v := GetLogContext(m.Ctx); v != nil {
		self.queue_write.Add(1)
		n, err = v.WriteLog(m)
	}
	return
}

func (self *WriterContext_t) LogRead(limit int) (out []Msg_t, ok bool) {
	return
}

func (self *WriterContext_t) WriteError(n int) {

}

func (self *WriterContext_t) Size() (res QueueSize_t) {
	res.QueueWrite = int(self.queue_write.Load())
	return
}

func (self *WriterContext_t) WgAdd(int) {

}

func (self *WriterContext_t) WgDone() {

}

func (self *WriterContext_t) Close() error {
	return nil
}

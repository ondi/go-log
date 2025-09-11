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

	"github.com/google/uuid"
	"github.com/ondi/go-circular"
)

// &log_ctx used for ctx.Value
var log_ctx = 1

type RangeFn_t = func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool

type LogContext interface {
	ContextName() string
	WriteLog(m Msg_t) (n int, err error)
	ContextRange(f RangeFn_t)
	ContextReset()
}

func SetLogContext(ctx context.Context, value LogContext) context.Context {
	return context.WithValue(ctx, &log_ctx, value)
}

func GetLogContext(ctx context.Context) (value LogContext) {
	value, _ = ctx.Value(&log_ctx).(LogContext)
	return
}

type LogContext_t struct {
	mx    sync.Mutex
	name  string
	data  *circular.List_t[Msg_t]
	limit int
}

func NewLogContext(name string, limit int) (self *LogContext_t) {
	self = &LogContext_t{
		name:  name,
		limit: limit,
		data:  circular.New[Msg_t](limit),
	}
	return
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
		return f(m.Info.Ts, m.Info.File, m.Info.Line, m.Info.Level, m.Format, m.Args...)
	})
}

func (self *LogContext_t) ContextReset() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.Reset()
}

type LogContextRead_t struct {
	level int64
}

func NewLogContextRead(level int64) (self *LogContextRead_t) {
	self = &LogContextRead_t{
		level: level,
	}
	return
}

func (self *LogContextRead_t) GetPayload(ctx context.Context, out func(key string, value string, args ...any)) {
	if v := GetLogContext(ctx); v != nil {
		v.ContextRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			if level_id < self.level {
				return true
			}
			out(fmt.Sprintf("LEVEL%v", level_id), format, args)
			return false
		})
	}
}

type LogContextMiddleware_t struct {
	Handler http.Handler
	Limit   int
}

func NewLogContextMiddleware(next http.Handler, limit int) http.Handler {
	self := &LogContextMiddleware_t{
		Handler: next,
		Limit:   limit,
	}
	return self
}

func (self *LogContextMiddleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), &log_ctx, NewLogContext(uuid.New().String(), self.Limit)))
	self.Handler.ServeHTTP(w, r)
}

type LogContextWriter_t struct {
	queue_write atomic.Int64
}

func NewLogContextWriter() Queue {
	return &LogContextWriter_t{}
}

func (self *LogContextWriter_t) LogWrite(msg []Msg_t) (n int, err error) {
	for _, m := range msg {
		if v := GetLogContext(m.Ctx); v != nil {
			self.queue_write.Add(1)
			n, err = v.WriteLog(m)
		}
	}
	return
}

func (self *LogContextWriter_t) Size() (res QueueSize_t) {
	res.QueueWrite = int(self.queue_write.Load())
	return
}

func (self *LogContextWriter_t) Close() error {
	return nil
}

//
//
//

package log

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/ondi/go-circular"
)

// &log_circular used for ctx.Value
var log_circular = 1

type RangeFn_t = func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool

type LogCircular interface {
	CircularName() string
	WriteLog(m Msg_t) (n int, err error)
	CircularRange(f RangeFn_t)
	CircularReset()
}

func SetLogCircular(ctx context.Context, value LogCircular) context.Context {
	return context.WithValue(ctx, &log_circular, value)
}

func GetLogCircular(ctx context.Context) (value LogCircular) {
	value, _ = ctx.Value(&log_circular).(LogCircular)
	return
}

type LogCircular_t struct {
	mx    sync.Mutex
	name  string
	data  *circular.List_t[Msg_t]
	limit int
}

func NewLogCircular(name string, limit int) (self *LogCircular_t) {
	self = &LogCircular_t{
		name:  name,
		limit: limit,
		data:  circular.New[Msg_t](limit),
	}
	return
}

func (self *LogCircular_t) CircularName() string {
	return self.name
}

func (self *LogCircular_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.data.Size() >= self.limit {
		self.data.PopFront()
	}
	self.data.PushBack(m)
	return
}

func (self *LogCircular_t) CircularRange(f RangeFn_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.RangeFront(func(m Msg_t) bool {
		return f(m.Info.Ts, m.Info.File, m.Info.Line, m.Info.Level, m.Format, m.Args...)
	})
}

func (self *LogCircular_t) CircularReset() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.Reset()
}

type LogCircularMiddleware_t struct {
	Handler http.Handler
	Limit   int
}

func NewLogCircularMiddleware(next http.Handler, limit int) http.Handler {
	self := &LogCircularMiddleware_t{
		Handler: next,
		Limit:   limit,
	}
	return self
}

func (self *LogCircularMiddleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), &log_circular, NewLogCircular(uuid.New().String(), self.Limit)))
	self.Handler.ServeHTTP(w, r)
}

type LogCircularWriter_t struct {
	queue_write atomic.Int64
}

func NewLogCircularWriter() Queue {
	return &LogCircularWriter_t{}
}

func (self *LogCircularWriter_t) LogWrite(msg []Msg_t) (n int, err error) {
	for _, m := range msg {
		if v := GetLogCircular(m.Ctx); v != nil {
			self.queue_write.Add(1)
			n, err = v.WriteLog(m)
		}
	}
	return
}

func (self *LogCircularWriter_t) Size() (res QueueSize_t) {
	res.QueueWrite = int(self.queue_write.Load())
	return
}

func (self *LogCircularWriter_t) Close() error {
	return nil
}

type LogCircularRead_t struct {
	levels map[int64]bool
}

func NewLogCircularRead(levels ...int64) (self *LogCircularRead_t) {
	self = &LogCircularRead_t{
		levels: map[int64]bool{},
	}
	for _, v := range levels {
		self.levels[v] = true
	}
	return
}

func (self *LogCircularRead_t) GetPayload(ctx context.Context, out func(level_id int64, format string, args ...any)) {
	if v := GetLogCircular(ctx); v != nil {
		v.CircularRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			if self.levels[level_id] == false {
				return true
			}
			out(level_id, format, args)
			return false
		})
	}
}

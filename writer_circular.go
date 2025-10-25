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

type LogTag interface {
	LogTagKey() string
	LogTagValue() string
}

type LogTag_t struct {
	Key   string
	Value string
}

func (self LogTag_t) LogTagKey() string {
	return self.Key
}

func (self LogTag_t) LogTagValue() string {
	return self.Value
}

func (self LogTag_t) String() string {
	return self.Key + "=" + self.Value
}

type LogCircularRead_t struct{}

func NewLogCircularRead() (self *LogCircularRead_t) {
	return &LogCircularRead_t{}
}

func (self *LogCircularRead_t) GetAll(ctx context.Context, out func(level_id int64, format string, args ...any) bool) {
	if v := GetLogCircular(ctx); v != nil {
		v.CircularRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			return out(level_id, format, args)
		})
	}
}

func (self *LogCircularRead_t) GetLevels(ctx context.Context, out map[string]int64) {
	if v := GetLogCircular(ctx); v != nil {
		v.CircularRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			out[fmt.Sprintf("LEVEL%v", level_id)]++
			return true
		})
	}
}

func (self *LogCircularRead_t) GetTags(ctx context.Context, out map[string]string) {
	if v := GetLogCircular(ctx); v != nil {
		v.CircularRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			for _, v2 := range args {
				if temp, ok := v2.(LogTag); ok {
					out[temp.LogTagKey()] = temp.LogTagValue()
				}
			}
			return true
		})
	}
}

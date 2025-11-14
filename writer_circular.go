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
	CircularSet(key string, value string)
	CircularGet(key string) (value string)
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
	kv    map[string]string
	data  *circular.List_t[Msg_t]
	limit int
}

func NewLogCircular(id string, limit int) (self *LogCircular_t) {
	self = &LogCircular_t{
		kv:    map[string]string{"id": id},
		data:  circular.New[Msg_t](limit),
		limit: limit,
	}
	return
}

func (self *LogCircular_t) CircularSet(key string, value string) {
	self.mx.Lock()
	self.kv[key] = value
	self.mx.Unlock()
}

func (self *LogCircular_t) CircularGet(key string) (value string) {
	self.mx.Lock()
	value = self.kv[key]
	self.mx.Unlock()
	return
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

type Tag interface {
	TagKey() string
	TagValue() string
}

type Tag_t struct {
	Key   string
	Value string
}

func (self Tag_t) LogTagKey() string {
	return self.Key
}

func (self Tag_t) LogTagValue() string {
	return self.Value
}

func (self Tag_t) String() string {
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

func (self *LogCircularRead_t) CountTags(ctx context.Context, out map[string]int64) {
	if v := GetLogCircular(ctx); v != nil {
		v.CircularRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			out[LevelName(level_id)]++
			for _, v2 := range args {
				if temp, ok := v2.(Tag); ok {
					out[temp.TagKey()]++
				}
			}
			return true
		})
	}
}

func (self *LogCircularRead_t) GetTags(ctx context.Context, out map[string]string) {
	if v := GetLogCircular(ctx); v != nil {
		v.CircularRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			for _, v2 := range args {
				if temp, ok := v2.(Tag); ok {
					out[temp.TagKey()] = temp.TagValue()
				}
			}
			return true
		})
	}
}

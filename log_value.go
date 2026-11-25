//
//
//

package log

import (
	"context"
	"net/http"
	"sync"
	"uuid"
)

type log_key_t int64

var my_log_key log_key_t = 1

type LogValue interface {
	LogValueId() string
	SetValue(key string, value string)
	GetValue(key string) (value string)
	SetQueue(q Queue)
	GetQueue() Queue
	Range(f func(key, value string) bool)
}

func SetLogValue(ctx context.Context, in LogValue) (out context.Context) {
	out = context.WithValue(ctx, my_log_key, in)
	return
}

func GetLogValue(ctx context.Context) (out LogValue) {
	out, _ = ctx.Value(my_log_key).(LogValue)
	return
}

type LogValue_t struct {
	mx           sync.Mutex
	queue        Queue
	value        map[string]string
	log_value_id string
}

func NewLogValue(log_value_id string) (self *LogValue_t) {
	return &LogValue_t{
		value:        map[string]string{},
		log_value_id: log_value_id,
	}
}

func (self *LogValue_t) LogValueId() string {
	return self.log_value_id
}

func (self *LogValue_t) SetValue(key string, value string) {
	self.mx.Lock()
	self.value[key] = value
	self.mx.Unlock()
}

func (self *LogValue_t) GetValue(key string) (value string) {
	self.mx.Lock()
	value = self.value[key]
	self.mx.Unlock()
	return
}

func (self *LogValue_t) Range(f func(key, value string) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.value {
		if f(k, v) == false {
			return
		}
	}
}

func (self *LogValue_t) SetQueue(queue Queue) {
	self.mx.Lock()
	self.queue = queue
	self.mx.Unlock()
}

func (self *LogValue_t) GetQueue() (queue Queue) {
	self.mx.Lock()
	queue = self.queue
	self.mx.Unlock()
	return
}

type LogCounterMiddleware_t struct {
	handler http.Handler
}

func NewLogCounterMiddleware(next http.Handler) http.Handler {
	self := &LogCounterMiddleware_t{
		handler: next,
	}
	return self
}

func (self *LogCounterMiddleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log_value := NewLogValue(uuid.New().String())
	log_value.SetQueue(NewWriterCounter())
	r = r.WithContext(SetLogValue(r.Context(), log_value))
	self.handler.ServeHTTP(w, r)
}

func LogCounterMiddlewareReader(ctx context.Context, f func(name string, level string, key string, value int64) bool) {
	if l := GetLogValue(ctx); l != nil {
		LogCounterReader(l, f)
	}
}

func LogCounterReader(l LogValue, f func(name string, level string, key string, value int64) bool) {
	if q := l.GetQueue(); q != nil {
		if c, _ := q.(*WriterCounter_t); c != nil {
			c.Range(func(level string, key string, count int64) bool {
				return f("logs", level, key, count)
			})
		}
	}
	l.Range(func(key string, value string) bool {
		return f(key, "", value, 1)
	})
}

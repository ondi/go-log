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

type log_key_t int64

var LogKey = []log_key_t{
	0,
	1,
}

type RangeFn_t = func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool

type LogBuffer interface {
	BufferId() string
	ExtraTagsSet(name string, level string, key string, value string)
	ExtraTagsGetCount(f func(name string, level string, key string, count int64) bool)
	ExtraTagsGetAll(f func(name string, level string, key string, value string) bool)
	WriteLog(m Msg_t) (n int, err error)
	BufferRange(f RangeFn_t)
	BufferReset()
}

func SetLogBuffer(ctx context.Context, key log_key_t, value LogBuffer) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetLogBuffer(ctx context.Context, key log_key_t) (value LogBuffer) {
	value, _ = ctx.Value(key).(LogBuffer)
	return
}

type LogBuffer_t struct {
	mx         sync.Mutex
	buffer_id  string
	extra_tags map[string]map[string]map[string][]string
	data       *circular.List_t[Msg_t]
	limit      int
}

func NewLogBuffer(buffer_id string, limit int) (self *LogBuffer_t) {
	self = &LogBuffer_t{
		buffer_id:  buffer_id,
		extra_tags: map[string]map[string]map[string][]string{},
		data:       circular.New[Msg_t](limit),
		limit:      limit,
	}
	return
}

func (self *LogBuffer_t) BufferId() string {
	return self.buffer_id
}

func (self *LogBuffer_t) ExtraTagsSet(name string, level string, key string, value string) {
	self.mx.Lock()
	if self.extra_tags[name] == nil {
		self.extra_tags[name] = map[string]map[string][]string{}
	}
	if self.extra_tags[name][level] == nil {
		self.extra_tags[name][level] = map[string][]string{}
	}
	self.extra_tags[name][level][key] = append(self.extra_tags[name][level][key], value)
	self.mx.Unlock()
}

func (self *LogBuffer_t) ExtraTagsGetCount(f func(name string, level string, key string, count int64) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for name, v1 := range self.extra_tags {
		for level, v2 := range v1 {
			for key, v3 := range v2 {
				if f(name, level, key, int64(len(v3))) == false {
					return
				}
			}
		}
	}
}

func (self *LogBuffer_t) ExtraTagsGetAll(f func(name string, level string, key string, value string) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for name, v1 := range self.extra_tags {
		for level, v2 := range v1 {
			for key, v3 := range v2 {
				for _, value := range v3 {
					if f(name, level, key, value) == false {
						return
					}
				}
			}
		}
	}
}

func (self *LogBuffer_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.data.Size() >= self.limit {
		self.data.PopFront()
	}
	self.data.PushBack(m)
	return
}

func (self *LogBuffer_t) BufferRange(f RangeFn_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.RangeFront(func(m Msg_t) bool {
		return f(m.Info.Ts, m.Info.File, m.Info.Line, m.Info.Level, m.Format, m.Args...)
	})
}

func (self *LogBuffer_t) BufferReset() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.Reset()
}

type LogBufferWriter_t struct {
	queue_write atomic.Int64
}

func NewLogBufferWriter() Queue {
	return &LogBufferWriter_t{}
}

func (self *LogBufferWriter_t) LogWrite(msg []Msg_t) (n int, err error) {
	for _, m := range msg {
		for _, v := range LogKey {
			if v := GetLogBuffer(m.Ctx, v); v != nil {
				self.queue_write.Add(1)
				n, err = v.WriteLog(m)
			}
		}
	}
	return
}

func (self *LogBufferWriter_t) Size() (res QueueSize_t) {
	res.QueueWrite = int(self.queue_write.Load())
	return
}

func (self *LogBufferWriter_t) Close() error {
	return nil
}

type LogBufferReader_t struct {
	key log_key_t
}

func NewLogBufferReader(key log_key_t) (self *LogBufferReader_t) {
	return &LogBufferReader_t{
		key: key,
	}
}

func (self *LogBufferReader_t) GetMsg(ctx context.Context, out func(level_id int64, format string, args ...any) bool) {
	if v := GetLogBuffer(ctx, self.key); v != nil {
		v.BufferRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			return out(level_id, format, args)
		})
	}
}

func (self *LogBufferReader_t) GetTagsCount(ctx context.Context, f func(name string, level string, key string, count int64) bool) {
	if v := GetLogBuffer(ctx, self.key); v != nil {
		tags := map[string]map[string]int64{}
		v.BufferRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			found := 0
			level := LevelName(level_id)
			if tags[level] == nil {
				tags[level] = map[string]int64{}
			}
			for _, v2 := range args {
				if temp, ok := v2.(Tag); ok {
					tags[level][temp.TagKey()]++
					found++
				}
			}
			if found == 0 {
				tags[level][""]++
			}
			return true
		})
		for level, v1 := range tags {
			for key, v2 := range v1 {
				f("logs", level, key, v2)
			}
		}
		v.ExtraTagsGetCount(f)
	}
}

func (self *LogBufferReader_t) GetTagsAll(ctx context.Context, f func(name string, level string, key string, value string) bool) {
	if v := GetLogBuffer(ctx, self.key); v != nil {
		tags := map[string]map[string]string{}
		v.BufferRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			found := 0
			level := LevelName(level_id)
			if tags[level] == nil {
				tags[level] = map[string]string{}
			}
			for _, v2 := range args {
				if temp, ok := v2.(Tag); ok {
					tags[level][temp.TagKey()] = temp.TagValue()
					found++
				}
			}
			if found == 0 {
				tags[level][""] = ""
			}
			return true
		})
		for level, v1 := range tags {
			for key, v2 := range v1 {
				f("logs", level, key, v2)
			}
		}
		v.ExtraTagsGetAll(f)
	}
}

type LogBufferMiddleware_t struct {
	handler http.Handler
	key     log_key_t
	limit   int
}

func NewLogBufferMiddleware(next http.Handler, key log_key_t, limit int) http.Handler {
	self := &LogBufferMiddleware_t{
		handler: next,
		key:     key,
		limit:   limit,
	}
	return self
}

func (self *LogBufferMiddleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(SetLogBuffer(r.Context(), self.key, NewLogBuffer(uuid.New().String(), self.limit)))
	self.handler.ServeHTTP(w, r)
}

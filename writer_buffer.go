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

// &log_buffer used for ctx.Value
var log_buffer = 1

type RangeFn_t = func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool

type LogBuffer interface {
	BufferId() string
	ExtraTagsSet(level string, key string, value string)
	ExtraTagsGetAll(out map[string]map[string]string)
	ExtraTagsGetCount(out map[string]map[string]int64)
	WriteLog(m Msg_t) (n int, err error)
	BufferRange(f RangeFn_t)
	BufferReset()
}

func SetLogBuffer(ctx context.Context, value LogBuffer) context.Context {
	return context.WithValue(ctx, &log_buffer, value)
}

func GetLogBuffer(ctx context.Context) (value LogBuffer) {
	value, _ = ctx.Value(&log_buffer).(LogBuffer)
	return
}

type LogBuffer_t struct {
	mx         sync.Mutex
	buffer_id  string
	extra_tags map[string]map[string]string
	data       *circular.List_t[Msg_t]
	limit      int
}

func NewLogBuffer(buffer_id string, limit int) (self *LogBuffer_t) {
	self = &LogBuffer_t{
		buffer_id:  buffer_id,
		extra_tags: map[string]map[string]string{},
		data:       circular.New[Msg_t](limit),
		limit:      limit,
	}
	return
}

func (self *LogBuffer_t) BufferId() string {
	return self.buffer_id
}

func (self *LogBuffer_t) ExtraTagsSet(level string, key string, value string) {
	self.mx.Lock()
	if self.extra_tags[level] == nil {
		self.extra_tags[level] = map[string]string{}
	}
	self.extra_tags[level][key] = value
	self.mx.Unlock()
}

func (self *LogBuffer_t) ExtraTagsGetAll(out map[string]map[string]string) {
	self.mx.Lock()
	for k1, v1 := range self.extra_tags {
		for k2, v2 := range v1 {
			if out[k1] == nil {
				out[k1] = map[string]string{}
			}
			out[k1][k2] = v2
		}
	}
	self.mx.Unlock()
}

func (self *LogBuffer_t) ExtraTagsGetCount(out map[string]map[string]int64) {
	self.mx.Lock()
	for k1, v1 := range self.extra_tags {
		for k2 := range v1 {
			if out[k1] == nil {
				out[k1] = map[string]int64{}
			}
			out[k1][k2]++
		}
	}
	self.mx.Unlock()
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

type LogBufferMiddleware_t struct {
	Handler http.Handler
	Limit   int
}

func NewLogBufferMiddleware(next http.Handler, limit int) http.Handler {
	self := &LogBufferMiddleware_t{
		Handler: next,
		Limit:   limit,
	}
	return self
}

func (self *LogBufferMiddleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(SetLogBuffer(r.Context(), NewLogBuffer(uuid.New().String(), self.Limit)))
	self.Handler.ServeHTTP(w, r)
}

type LogBufferWriter_t struct {
	queue_write atomic.Int64
}

func NewLogBufferWriter() Queue {
	return &LogBufferWriter_t{}
}

func (self *LogBufferWriter_t) LogWrite(msg []Msg_t) (n int, err error) {
	for _, m := range msg {
		if v := GetLogBuffer(m.Ctx); v != nil {
			self.queue_write.Add(1)
			n, err = v.WriteLog(m)
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

type LogBufferRead_t struct{}

func NewLogBufferRead() (self *LogBufferRead_t) {
	return &LogBufferRead_t{}
}

func (self *LogBufferRead_t) GetMsg(ctx context.Context, out func(level_id int64, format string, args ...any) bool) {
	if v := GetLogBuffer(ctx); v != nil {
		v.BufferRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			return out(level_id, format, args)
		})
	}
}

func (self *LogBufferRead_t) GetTagsCount(ctx context.Context, out map[string]map[string]int64) {
	if v := GetLogBuffer(ctx); v != nil {
		v.BufferRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			found := 0
			level := LevelName(level_id)
			if out[level] == nil {
				out[level] = map[string]int64{}
			}
			for _, v2 := range args {
				if temp, ok := v2.(Tag); ok {
					out[level][temp.TagKey()]++
					found++
				}
			}
			if found == 0 {
				out[level][""]++
			}
			return true
		})
		v.ExtraTagsGetCount(out)
	}
}

func (self *LogBufferRead_t) GetTagsAll(ctx context.Context, out map[string]map[string]string) {
	if v := GetLogBuffer(ctx); v != nil {
		v.BufferRange(func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool {
			found := 0
			level := LevelName(level_id)
			if out[level] == nil {
				out[level] = map[string]string{}
			}
			for _, v2 := range args {
				if temp, ok := v2.(Tag); ok {
					out[level][temp.TagKey()] = temp.TagValue()
					found++
				}
			}
			if found == 0 {
				out[level][""] = ""
			}
			return true
		})
		v.ExtraTagsGetAll(out)
	}
}

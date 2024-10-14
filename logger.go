/*
	Log(levels) with no allocation and locks
*/

package log

import (
	"context"
	"io"
	"sync/atomic"
	"time"
)

type Info_t struct {
	Ts        time.Time `json:"ts"`
	LevelName string    `json:"level_name"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	LevelId   int64     `json:"level"`
}

func (self *Info_t) Set(ts time.Time) {
	self.Ts = ts
	self.File, self.Line = FileLine(1, 32)
}

type Msg_t struct {
	Ctx    context.Context `json:"-"`
	Info   Info_t          `json:"info"`
	Format string          `json:"format"`
	Args   []any           `json:"args"`
}

type QueueSize_t struct {
	Limit      int
	Size       int
	Readers    int
	Writers    int
	QueueWrite int
	QueueRead  int
	QueueError int
	WriteError int
}

type Queue interface {
	LogWrite(m Msg_t) (int, error)
	LogRead(p []Msg_t) (n int, ok bool)
	Size() QueueSize_t
	Close() error
	WgAdd(int)
	WgDone()
	WriteError(n int)
}

type Formatter interface {
	FormatMessage(out io.Writer, in ...Msg_t) (int, error)
}

type Logger interface {
	Trace(format string, args ...any)
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)

	TraceCtx(ctx context.Context, format string, args ...any)
	DebugCtx(ctx context.Context, format string, args ...any)
	InfoCtx(ctx context.Context, format string, args ...any)
	WarnCtx(ctx context.Context, format string, args ...any)
	ErrorCtx(ctx context.Context, format string, args ...any)

	Log(ctx context.Context, level Info_t, format string, args ...any)

	AddOutput(writer_name string, writer Queue, levels []Info_t) Logger
	DelOutput(writer_name string, levels []Info_t) Logger
	Range(fn func(level_id int64, writer_name string, writer Queue) bool)
	Close() Logger
}

type queue_map_t map[string]Queue

type queue_ptr_t struct {
	ptr atomic.Pointer[queue_map_t]
}

func (self *queue_ptr_t) create(writer_name string, new_writer Queue) (old_writer Queue) {
	for {
		old := self.ptr.Load()
		new := queue_map_t{}
		for k, v := range *old {
			if k == writer_name {
				old_writer = v
			} else {
				new[k] = v
			}
		}
		if new_writer != nil {
			new[writer_name] = new_writer
		}
		if self.ptr.CompareAndSwap(old, &new) {
			return
		}
	}
}

type level_map_t map[int64]*queue_ptr_t

type level_ptr_t struct {
	ptr atomic.Pointer[level_map_t]
}

func (self *level_ptr_t) create(level_id int64, writer_name string, writer Queue) {
	for {
		old := self.ptr.Load()
		level, ok := (*old)[level_id]
		if ok {
			level.create(writer_name, writer)
			return
		}
		new := level_map_t{}
		for k, v := range *old {
			new[k] = v
		}
		level = &queue_ptr_t{}
		level.ptr.Store(&queue_map_t{})
		level.create(writer_name, writer)
		new[level_id] = level
		if self.ptr.CompareAndSwap(old, &new) {
			return
		}
	}
}

type log_t struct {
	level_map level_ptr_t
}

func New() Logger {
	self := &log_t{}
	self.level_map.ptr.Store(&level_map_t{})
	return self
}

func (self *log_t) AddOutput(writer_name string, writer Queue, levels []Info_t) Logger {
	for _, level := range levels {
		self.level_map.create(level.LevelId, writer_name, writer)
	}
	return self
}

func (self *log_t) DelOutput(writer_name string, levels []Info_t) Logger {
	for _, v := range levels {
		if level := (*self.level_map.ptr.Load())[v.LevelId]; level != nil {
			if writer := level.create(writer_name, nil); writer != nil {
				writer.Close()
			}
		}
	}
	return self
}

func (self *log_t) Range(fn func(level_id int64, writer_name string, writer Queue) bool) {
	for level_id, level := range *self.level_map.ptr.Load() {
		for writer_name, writer := range *level.ptr.Load() {
			if fn(level_id, writer_name, writer) == false {
				return
			}
		}
	}
}

func (self *log_t) Close() Logger {
	for _, level := range *self.level_map.ptr.Load() {
		for _, writer := range *level.ptr.Swap(&queue_map_t{}) {
			writer.Close()
		}
	}
	return self
}

func (self *log_t) Log(ctx context.Context, info Info_t, format string, args ...any) {
	info.Set(time.Now())
	if level := (*self.level_map.ptr.Load())[info.LevelId]; level != nil {
		for _, writer := range *level.ptr.Load() {
			writer.LogWrite(Msg_t{Ctx: ctx, Info: info, Format: format, Args: args})
		}
	}
}

func (self *log_t) Error(format string, args ...any) {
	self.Log(context.Background(), LOG_ERROR, format, args...)
}

func (self *log_t) Warn(format string, args ...any) {
	self.Log(context.Background(), LOG_WARN, format, args...)
}

func (self *log_t) Info(format string, args ...any) {
	self.Log(context.Background(), LOG_INFO, format, args...)
}

func (self *log_t) Debug(format string, args ...any) {
	self.Log(context.Background(), LOG_DEBUG, format, args...)
}

func (self *log_t) Trace(format string, args ...any) {
	self.Log(context.Background(), LOG_TRACE, format, args...)
}

func (self *log_t) ErrorCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_ERROR, format, args...)
}

func (self *log_t) WarnCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_WARN, format, args...)
}

func (self *log_t) InfoCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_INFO, format, args...)
}

func (self *log_t) DebugCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_DEBUG, format, args...)
}

func (self *log_t) TraceCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_TRACE, format, args...)
}

func Error(format string, args ...any) {
	__std.Error(format, args...)
}

func Warn(format string, args ...any) {
	__std.Warn(format, args...)
}

func Info(format string, args ...any) {
	__std.Info(format, args...)
}

func Debug(format string, args ...any) {
	__std.Debug(format, args...)
}

func Trace(format string, args ...any) {
	__std.Trace(format, args...)
}

func ErrorCtx(ctx context.Context, format string, args ...any) {
	__std.ErrorCtx(ctx, format, args...)
}

func WarnCtx(ctx context.Context, format string, args ...any) {
	__std.WarnCtx(ctx, format, args...)
}

func InfoCtx(ctx context.Context, format string, args ...any) {
	__std.InfoCtx(ctx, format, args...)
}

func DebugCtx(ctx context.Context, format string, args ...any) {
	__std.DebugCtx(ctx, format, args...)
}

func TraceCtx(ctx context.Context, format string, args ...any) {
	__std.TraceCtx(ctx, format, args...)
}

func SetLogger(in Logger) Logger {
	__std = in
	return __std
}

func GetLogger() Logger {
	return __std
}

/*
	Log with levels

	// no allocation and locks for WriteLog cycle
	func (self *log_t) Log(ctx context.Context, level Level_t, format string, args ...any) {
		level.Set(time.Now())
		if v1 := self.levels[level.Level]; v1 != nil {
			for _, v2 := range *v1.Load() {
				v2.WriteLog(Msg_t{Ctx: ctx, Level: level, Format: format, Args: args})
			}
		}
	}
*/

package log

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

type Level_t struct {
	Name  string
	File  string
	Line  int
	Level int64
	Ts    time.Time
}

func (self *Level_t) Set(ts time.Time) {
	self.Ts = ts
	self.File, self.Line = FileLine(1, 32)
}

type Msg_t struct {
	Ctx    context.Context
	Level  Level_t
	Format string
	Args   []any
}

type QueueSize_t struct {
	Limit   int
	Size    int
	Readers int
	Writers int
}

type Queue interface {
	WriteLog(m Msg_t) (int, error)
	ReadLog(count int) (out []Msg_t, oki int)
	Size() QueueSize_t
	Close() error
}

type Formatter interface {
	FormatLog(out io.Writer, m Msg_t) (int, error)
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

	Log(ctx context.Context, level Level_t, format string, args ...any)

	AddOutput(name string, writer Queue, in []Level_t) Logger
	DelOutput(name string) Logger
	RangeLevel(level Level_t, fn func(name string, queue Queue) bool)
	Clear() Logger
}

type writers_t map[string]Queue

func add_output(value *atomic.Pointer[writers_t], name string, writer Queue) {
	for {
		old := value.Load()
		new := writers_t{}
		for k, v := range *old {
			new[k] = v
		}
		new[name] = writer
		if value.CompareAndSwap(old, &new) {
			return
		}
	}
}

func del_output(value *atomic.Pointer[writers_t], name string) {
	for {
		old := value.Load()
		new := writers_t{}
		var writer Queue
		for k, v := range *old {
			if k == name {
				writer = v
			} else {
				new[k] = v
			}
		}
		if value.CompareAndSwap(old, &new) {
			if writer != nil {
				writer.Close()
			}
			return
		}
	}
}

type log_t struct {
	levels map[int64]*atomic.Pointer[writers_t]
}

func New(in []Level_t) Logger {
	self := &log_t{
		levels: map[int64]*atomic.Pointer[writers_t]{},
	}
	for _, v := range in {
		self.levels[v.Level] = &atomic.Pointer[writers_t]{}
		self.levels[v.Level].Store(&writers_t{})
	}
	return self
}

func (self *log_t) AddOutput(name string, writer Queue, in []Level_t) Logger {
	for _, v1 := range in {
		if v2 := self.levels[v1.Level]; v2 != nil {
			add_output(v2, name, writer)
		}
	}
	return self
}

func (self *log_t) DelOutput(name string) Logger {
	for _, v := range self.levels {
		del_output(v, name)
	}
	return self
}

func (self *log_t) RangeLevel(level Level_t, fn func(name string, queue Queue) bool) {
	if v1 := self.levels[level.Level]; v1 != nil {
		for k2, v2 := range *v1.Load() {
			if fn(k2, v2) == false {
				return
			}
		}
	}
}

func (self *log_t) Clear() Logger {
	for _, v1 := range self.levels {
		for _, v2 := range *v1.Swap(&writers_t{}) {
			v2.Close()
		}
	}
	return self
}

func (self *log_t) Log(ctx context.Context, level Level_t, format string, args ...any) {
	level.Set(time.Now())
	if v1 := self.levels[level.Level]; v1 != nil {
		for writer, v2 := range *v1.Load() {
			if _, err := v2.WriteLog(Msg_t{Ctx: ctx, Level: level, Format: format, Args: args}); err != nil {
				fmt.Fprintf(STDERR, "LOG ERROR: %v %v %v\n", level.Ts.Format("2006-01-02 15:04:05"), writer, err)
			}
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

/*
	Log with levels

	// no allocation and locks for WriteLog cycle
	func (self *log_t) Debug(format string, args ...any) {
		ts := time.Now()
		for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_DEBUG.Level])) {
			v.WriteLog(context.Background(), ts, LOG_DEBUG.Name, format, args...)
		}
	}
*/

package log

import (
	"context"
	"io"
	"sync/atomic"
	"time"
	"unsafe"
)

type Level_t struct {
	Name  string
	Level int64
}

var (
	LOG_TRACE = Level_t{Name: "TRACE", Level: 0}
	LOG_DEBUG = Level_t{Name: "DEBUG", Level: 1}
	LOG_INFO  = Level_t{Name: "INFO", Level: 2}
	LOG_WARN  = Level_t{Name: "WARN", Level: 3}
	LOG_ERROR = Level_t{Name: "ERROR", Level: 4}
)

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

	Clear() Logger
	AddOutput(name string, writer Writer, levels []Level_t) Logger
	DelOutput(name string) Logger
}

type Writer interface {
	WriteLog(ctx context.Context, ts time.Time, level string, format string, args ...any) (int, error)
	Close() error
}

type Formatter interface {
	FormatLog(ctx context.Context, out io.Writer, ts time.Time, level string, format string, args ...any) (int, error)
}

type writers_t map[string]Writer

func add_output(value *unsafe.Pointer, name string, writer Writer) {
	for {
		temp := writers_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(*writers_t)(p) {
			temp[k] = v
		}
		temp[name] = writer
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&temp)) {
			return
		}
	}
}

func del_output(value *unsafe.Pointer, name string) {
	for {
		var writer Writer
		temp := writers_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(*writers_t)(p) {
			if k == name {
				writer = v
			} else {
				temp[k] = v
			}
		}
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&temp)) {
			if writer != nil {
				writer.Close()
			}
			return
		}
	}
}

type log_t struct {
	levels []unsafe.Pointer
}

func New() Logger {
	self := &log_t{
		levels: make([]unsafe.Pointer, LOG_ERROR.Level+1),
	}
	self.Clear()
	return self
}

func (self *log_t) Clear() Logger {
	for i := 0; i < len(self.levels); i++ {
		atomic.StorePointer(&self.levels[i], unsafe.Pointer(&writers_t{}))
	}
	return self
}

func (self *log_t) AddOutput(name string, writer Writer, levels []Level_t) Logger {
	for _, v := range levels {
		add_output(&self.levels[v.Level], name, writer)
	}
	return self
}

func (self *log_t) DelOutput(name string) Logger {
	for i := range self.levels {
		del_output(&self.levels[i], name)
	}
	return self
}

func (self *log_t) Log(ctx context.Context, level Level_t, format string, args ...any) {
	ts := time.Now()
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.levels[level.Level])) {
		v.WriteLog(ctx, ts, level.Name, format, args...)
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
	std.Error(format, args...)
}

func Warn(format string, args ...any) {
	std.Warn(format, args...)
}

func Info(format string, args ...any) {
	std.Info(format, args...)
}

func Debug(format string, args ...any) {
	std.Debug(format, args...)
}

func Trace(format string, args ...any) {
	std.Trace(format, args...)
}

func ErrorCtx(ctx context.Context, format string, args ...any) {
	std.ErrorCtx(ctx, format, args...)
}

func WarnCtx(ctx context.Context, format string, args ...any) {
	std.WarnCtx(ctx, format, args...)
}

func InfoCtx(ctx context.Context, format string, args ...any) {
	std.InfoCtx(ctx, format, args...)
}

func DebugCtx(ctx context.Context, format string, args ...any) {
	std.DebugCtx(ctx, format, args...)
}

func TraceCtx(ctx context.Context, format string, args ...any) {
	std.TraceCtx(ctx, format, args...)
}

func SetLogger(logger Logger) {
	std = logger
}

func GetLogger() Logger {
	return std
}

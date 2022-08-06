/*
	Log with levels

	// no allocation and locks
	func (self *log_t) Debug(format string, args ...interface{}) {
		for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_DEBUG.level])) {
			v.WriteLevel(LOG_DEBUG.Name, format, args...)
		}
	}
*/

package log

import (
	"context"
	"sync/atomic"
	"unsafe"
)

type level_t int

type level_name_t struct {
	Name   string
	Levels []level_t
}

var LOG_TRACE = level_name_t{Name: "TRACE", Levels: []level_t{0, 1, 2, 3, 4}}
var LOG_DEBUG = level_name_t{Name: "DEBUG", Levels: []level_t{1, 2, 3, 4}}
var LOG_INFO = level_name_t{Name: "INFO", Levels: []level_t{2, 3, 4}}
var LOG_WARN = level_name_t{Name: "WARN", Levels: []level_t{3, 4}}
var LOG_ERROR = level_name_t{Name: "ERROR", Levels: []level_t{4}}

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})

	TraceCtx(ctx context.Context, format string, args ...interface{})
	DebugCtx(ctx context.Context, format string, args ...interface{})
	InfoCtx(ctx context.Context, format string, args ...interface{})
	WarnCtx(ctx context.Context, format string, args ...interface{})
	ErrorCtx(ctx context.Context, format string, args ...interface{})

	AddOutput(name string, writer Writer, level []level_t)
	DelOutput(name string)
	Clear()
}

type Writer interface {
	WriteLevel(level string, format string, args ...interface{}) (int, error)
	Close() error
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
	out [5]unsafe.Pointer
}

func NewEmpty() (self Logger) {
	self = &log_t{}
	self.Clear()
	return
}

func NewLogger(name string, writer Writer, level []level_t) (self Logger) {
	self = NewEmpty()
	self.AddOutput(name, writer, level)
	return
}

func (self *log_t) AddOutput(name string, writer Writer, level []level_t) {
	for _, v := range level {
		add_output(&self.out[v], name, writer)
	}
}

func (self *log_t) DelOutput(name string) {
	for i := 0; i < len(self.out); i++ {
		del_output(&self.out[i], name)
	}
}

func (self *log_t) Clear() {
	for i := 0; i < len(self.out); i++ {
		atomic.StorePointer(&self.out[i], unsafe.Pointer(&writers_t{}))
	}
}

func (self *log_t) Error(format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_ERROR.Levels[0]])) {
		v.WriteLevel(LOG_ERROR.Name, format, args...)
	}
}

func (self *log_t) Warn(format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_WARN.Levels[0]])) {
		v.WriteLevel(LOG_WARN.Name, format, args...)
	}
}

func (self *log_t) Info(format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_INFO.Levels[0]])) {
		v.WriteLevel(LOG_INFO.Name, format, args...)
	}
}

func (self *log_t) Debug(format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_DEBUG.Levels[0]])) {
		v.WriteLevel(LOG_DEBUG.Name, format, args...)
	}
}

func (self *log_t) Trace(format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_TRACE.Levels[0]])) {
		v.WriteLevel(LOG_TRACE.Name, format, args...)
	}
}

func (self *log_t) ErrorCtx(ctx context.Context, format string, args ...interface{}) {
	level := ContextStore(ctx, LOG_ERROR.Name, format, args...)
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_ERROR.Levels[0]])) {
		v.WriteLevel(level, format, args...)
	}
}

func (self *log_t) WarnCtx(ctx context.Context, format string, args ...interface{}) {
	level := ContextStore(ctx, LOG_WARN.Name, format, args...)
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_WARN.Levels[0]])) {
		v.WriteLevel(level, format, args...)
	}
}

func (self *log_t) InfoCtx(ctx context.Context, format string, args ...interface{}) {
	level := ContextStore(ctx, LOG_INFO.Name, format, args...)
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_INFO.Levels[0]])) {
		v.WriteLevel(level, format, args...)
	}
}

func (self *log_t) DebugCtx(ctx context.Context, format string, args ...interface{}) {
	level := ContextStore(ctx, LOG_DEBUG.Name, format, args...)
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_DEBUG.Levels[0]])) {
		v.WriteLevel(level, format, args...)
	}
}

func (self *log_t) TraceCtx(ctx context.Context, format string, args ...interface{}) {
	level := ContextStore(ctx, LOG_TRACE.Name, format, args...)
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[LOG_TRACE.Levels[0]])) {
		v.WriteLevel(level, format, args...)
	}
}

func Error(format string, args ...interface{}) {
	std.Error(format, args...)
}

func Warn(format string, args ...interface{}) {
	std.Warn(format, args...)
}

func Info(format string, args ...interface{}) {
	std.Info(format, args...)
}

func Debug(format string, args ...interface{}) {
	std.Debug(format, args...)
}

func Trace(format string, args ...interface{}) {
	std.Trace(format, args...)
}

func ErrorCtx(ctx context.Context, format string, args ...interface{}) {
	std.ErrorCtx(ctx, format, args...)
}

func WarnCtx(ctx context.Context, format string, args ...interface{}) {
	std.WarnCtx(ctx, format, args...)
}

func InfoCtx(ctx context.Context, format string, args ...interface{}) {
	std.InfoCtx(ctx, format, args...)
}

func DebugCtx(ctx context.Context, format string, args ...interface{}) {
	std.DebugCtx(ctx, format, args...)
}

func TraceCtx(ctx context.Context, format string, args ...interface{}) {
	std.TraceCtx(ctx, format, args...)
}

func SetLogger(logger Logger) {
	std = logger
}

func GetLogger() Logger {
	return std
}

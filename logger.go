/*
	Log with levels

	// the most used range cycle will not allocate new map for writers list
	func (self *log_t) Error(format string, args ...interface{}) {
		for _, v := range *(*writers_t)(atomic.LoadPointer(&self.err)) {
			v.WriteLevel(Level_t{"ERROR"}, format, args...)
		}
	}
*/

package log

import (
	"sync/atomic"
	"unsafe"
)

var LOG_TRACE = levels_t{Name: "TRACE", level: 0}
var LOG_DEBUG = levels_t{Name: "DEBUG", level: 1}
var LOG_INFO = levels_t{Name: "INFO", level: 2}
var LOG_WARN = levels_t{Name: "WARN", level: 3}
var LOG_ERROR = levels_t{Name: "ERROR", level: 4}

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	WriteLevel(level Levels, format string, args ...interface{})

	AddOutput(name string, level Levels, writer Writer)
	DelOutput(name string)
	Clear()
}

type Levels interface {
	String() string
	Level() int
}

type Writer interface {
	WriteLevel(level Levels, format string, args ...interface{}) (int, error)
	Close() error
}

type writers_t map[string]Writer

type levels_t struct {
	Name  string
	level int
}

func (self levels_t) String() string {
	return self.Name
}

func (self levels_t) Level() int {
	return self.level
}

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

func NewLogger(name string, level Levels, writer Writer) (self Logger) {
	self = NewEmpty()
	self.AddOutput(name, level, writer)
	return
}

func (self *log_t) AddOutput(name string, level Levels, writer Writer) {
	if level.Level() <= LOG_ERROR.level {
		add_output(&self.out[LOG_ERROR.level], name, writer)
	}
	if level.Level() <= LOG_WARN.level {
		add_output(&self.out[LOG_WARN.level], name, writer)
	}
	if level.Level() <= LOG_INFO.level {
		add_output(&self.out[LOG_INFO.level], name, writer)
	}
	if level.Level() <= LOG_DEBUG.level {
		add_output(&self.out[LOG_DEBUG.level], name, writer)
	}
	if level.Level() <= LOG_TRACE.level {
		add_output(&self.out[LOG_TRACE.level], name, writer)
	}
}

func (self *log_t) DelOutput(name string) {
	del_output(&self.out[LOG_ERROR.level], name)
	del_output(&self.out[LOG_WARN.level], name)
	del_output(&self.out[LOG_INFO.level], name)
	del_output(&self.out[LOG_DEBUG.level], name)
	del_output(&self.out[LOG_TRACE.level], name)
}

func (self *log_t) Clear() {
	atomic.StorePointer(&self.out[LOG_ERROR.level], unsafe.Pointer(&writers_t{}))
	atomic.StorePointer(&self.out[LOG_WARN.level], unsafe.Pointer(&writers_t{}))
	atomic.StorePointer(&self.out[LOG_INFO.level], unsafe.Pointer(&writers_t{}))
	atomic.StorePointer(&self.out[LOG_DEBUG.level], unsafe.Pointer(&writers_t{}))
	atomic.StorePointer(&self.out[LOG_TRACE.level], unsafe.Pointer(&writers_t{}))
}

func (self *log_t) Error(format string, args ...interface{}) {
	self.WriteLevel(LOG_ERROR, format, args...)
}

func (self *log_t) Warn(format string, args ...interface{}) {
	self.WriteLevel(LOG_WARN, format, args...)
}

func (self *log_t) Info(format string, args ...interface{}) {
	self.WriteLevel(LOG_INFO, format, args...)
}

func (self *log_t) Debug(format string, args ...interface{}) {
	self.WriteLevel(LOG_DEBUG, format, args...)
}

func (self *log_t) Trace(format string, args ...interface{}) {
	self.WriteLevel(LOG_TRACE, format, args...)
}

func (self *log_t) WriteLevel(level Levels, format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[level.Level()])) {
		v.WriteLevel(level, format, args...)
	}
}

func Trace(format string, args ...interface{}) {
	std.Trace(format, args...)
}

func Debug(format string, args ...interface{}) {
	std.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	std.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	std.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	std.Error(format, args...)
}

func WriteLevel(level Levels, format string, args ...interface{}) {
	std.WriteLevel(level, format, args...)
}

func SetLogger(logger Logger) {
	std = logger
}

func GetLogger() Logger {
	return std
}

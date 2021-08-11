/*
	Log with levels

	// no allocation and locks
	func (self *log_t) WriteLevel(level Levels, format string, args ...interface{}) {
		for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[level.level])) {
			v.WriteLevel(level.Name, format, args...)
		}
	}
*/

package log

import (
	"sync/atomic"
	"unsafe"
)

type level_t struct {
	Name  string
	level int
}

var LOG_TRACE = level_t{Name: "TRACE", level: 0}
var LOG_DEBUG = level_t{Name: "DEBUG", level: 1}
var LOG_INFO = level_t{Name: "INFO", level: 2}
var LOG_WARN = level_t{Name: "WARN", level: 3}
var LOG_ERROR = level_t{Name: "ERROR", level: 4}

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	WriteLevel(level level_t, format string, args ...interface{})

	AddOutput(name string, level level_t, writer Writer)
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

func NewLogger(name string, level level_t, writer Writer) (self Logger) {
	self = NewEmpty()
	self.AddOutput(name, level, writer)
	return
}

func (self *log_t) AddOutput(name string, level level_t, writer Writer) {
	for ; level.level < len(self.out); level.level++ {
		add_output(&self.out[level.level], name, writer)
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

func (self *log_t) WriteLevel(level level_t, format string, args ...interface{}) {
	for _, v := range *(*writers_t)(atomic.LoadPointer(&self.out[level.level])) {
		v.WriteLevel(level.Name, format, args...)
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

func WriteLevel(level level_t, format string, args ...interface{}) {
	std.WriteLevel(level, format, args...)
}

func SetLogger(logger Logger) {
	std = logger
}

func GetLogger() Logger {
	return std
}

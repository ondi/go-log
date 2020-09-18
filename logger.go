//
// Log with levels
//
// most used get_output() function will not allocate new slice or map for writers list
//

package log

import (
	"sync/atomic"
	"unsafe"
)

const (
	LOG_TRACE = 0
	LOG_DEBUG = 1
	LOG_INFO  = 2
	LOG_WARN  = 3
	LOG_ERROR = 4
)

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})

	AddOutput(name string, level int, writer Writer)
	DelOutput(name string)
	Clear()
}

type Writer interface {
	WriteLevel(level string, format string, args ...interface{}) (int, error)
}

type Prefixer interface {
	Prefix() string
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
		temp := writers_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(*writers_t)(p) {
			temp[k] = v
		}
		delete(temp, name)
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&temp)) {
			return
		}
	}
}

func clear_output(value *unsafe.Pointer) {
	atomic.StorePointer(value, unsafe.Pointer(&writers_t{}))
}

func get_output(value *unsafe.Pointer) writers_t {
	return *(*writers_t)(atomic.LoadPointer(value))
}

type log_t struct {
	err   unsafe.Pointer
	warn  unsafe.Pointer
	info  unsafe.Pointer
	debug unsafe.Pointer
	trace unsafe.Pointer
}

func NewEmpty() (self Logger) {
	self = &log_t{}
	self.Clear()
	return
}

func NewLogger(name string, level int, writer Writer) (self Logger) {
	self = NewEmpty()
	self.AddOutput(name, level, writer)
	return
}

func (self *log_t) AddOutput(name string, level int, writer Writer) {
	if level <= LOG_ERROR {
		add_output(&self.err, name, writer)
	}
	if level <= LOG_WARN {
		add_output(&self.warn, name, writer)
	}
	if level <= LOG_INFO {
		add_output(&self.info, name, writer)
	}
	if level <= LOG_DEBUG {
		add_output(&self.debug, name, writer)
	}
	if level <= LOG_TRACE {
		add_output(&self.trace, name, writer)
	}
}

func (self *log_t) DelOutput(name string) {
	del_output(&self.err, name)
	del_output(&self.warn, name)
	del_output(&self.info, name)
	del_output(&self.debug, name)
	del_output(&self.trace, name)
}

func (self *log_t) Clear() {
	clear_output(&self.err)
	clear_output(&self.warn)
	clear_output(&self.info)
	clear_output(&self.debug)
	clear_output(&self.trace)
}

func (self *log_t) Error(format string, args ...interface{}) {
	for _, v := range get_output(&self.err) {
		v.WriteLevel("ERROR", format, args...)
	}
}

func (self *log_t) Warn(format string, args ...interface{}) {
	for _, v := range get_output(&self.warn) {
		v.WriteLevel("WARN", format, args...)
	}
}

func (self *log_t) Info(format string, args ...interface{}) {
	for _, v := range get_output(&self.info) {
		v.WriteLevel("INFO", format, args...)
	}
}

func (self *log_t) Debug(format string, args ...interface{}) {
	for _, v := range get_output(&self.debug) {
		v.WriteLevel("DEBUG", format, args...)
	}
}

func (self *log_t) Trace(format string, args ...interface{}) {
	for _, v := range get_output(&self.trace) {
		v.WriteLevel("TRACE", format, args...)
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

func SetLogger(logger Logger) {
	std = logger
}

func GetLogger() Logger {
	return std
}

//
// Log with levels
//

package log

import "os"
import "io"
import "fmt"
import "log"
import "time"
import "sync/atomic"
import "unsafe"

const (
	LOG_TRACE = 0
	LOG_DEBUG = 1
	LOG_INFO = 2
	LOG_WARN = 3
	LOG_ERROR = 4
	
	DATETIME1 = "2006-01-02 15:04:05"
	DATETIME2 = "2006-01-02 15:04:05.000"
)

var std = NewLogger("stderr", LOG_TRACE, os.Stderr, DATETIME1)

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	
	AddOutput(name string, level int, out io.Writer, datetime string)
	DelOutput(name string)
	Clear()
}

type out_t struct {
	datetime func() string
	out io.Writer
}

type outmap_t map[string]out_t

func AddOutput(value * unsafe.Pointer, name string, out out_t) {
	for {
		temp := outmap_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(* outmap_t)(p) {
			temp[k] = v
		}
		temp[name] = out
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&temp)) {
			return
		}
	}
}

func DelOutput(value * unsafe.Pointer, name string) {
	for {
		out := outmap_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(* outmap_t)(p) {
			out[k] = v
		}
		delete(out, name)
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&out)) {
			return
		}
	}
}

func Range(value * unsafe.Pointer) outmap_t {
	return *(* outmap_t)(atomic.LoadPointer(value))
}

type LogLogger struct {
	error unsafe.Pointer
	warn unsafe.Pointer
	info unsafe.Pointer
	debug unsafe.Pointer
	trace unsafe.Pointer
}

func NewEmpty() Logger {
	self := &LogLogger{}
	self.Clear()
	return self
}

func NewLogger(name string, level int, out io.Writer, datetime string) Logger {
	self := NewEmpty()
	self.AddOutput(name, level, out, datetime)
	return self
}

func (self * LogLogger) AddOutput(name string, level int, out io.Writer, datetime string) {
	value := out_t{out: out}
	if len(datetime) > 0 {
		datetime += " "
		value.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		value.datetime = func() string {return ""}
	}
	if level <= LOG_ERROR {
		AddOutput(&self.error, name, value)
	}
	if level <= LOG_WARN {
		AddOutput(&self.warn, name, value)
	}
	if level <= LOG_INFO {
		AddOutput(&self.info, name, value)
	}
	if level <= LOG_DEBUG {
		AddOutput(&self.debug, name, value)
	}
	if level <= LOG_TRACE {
		AddOutput(&self.trace, name, value)
	}
}

func (self * LogLogger) DelOutput(name string) {
	DelOutput(&self.error, name)
	DelOutput(&self.warn, name)
	DelOutput(&self.info, name)
	DelOutput(&self.debug, name)
	DelOutput(&self.trace, name)
}

func (self * LogLogger) Clear() {
	atomic.StorePointer(&self.error, unsafe.Pointer(&outmap_t{}))
	atomic.StorePointer(&self.warn, unsafe.Pointer(&outmap_t{}))
	atomic.StorePointer(&self.info, unsafe.Pointer(&outmap_t{}))
	atomic.StorePointer(&self.debug, unsafe.Pointer(&outmap_t{}))
	atomic.StorePointer(&self.trace, unsafe.Pointer(&outmap_t{}))
}

func (self * LogLogger) Error(format string, args ...interface{}) {
	for _, v := range Range(&self.error) {
		fmt.Fprintf(v.out, v.datetime() + "ERROR " + format + "\n", args...)
	}
}

func (self * LogLogger) Warn(format string, args ...interface{}) {
	for _, v := range Range(&self.warn) {
		fmt.Fprintf(v.out, v.datetime() + "WARN " + format + "\n", args...)
	}
}

func (self * LogLogger) Info(format string, args ...interface{}) {
	for _, v := range Range(&self.info) {
		fmt.Fprintf(v.out, v.datetime() + "INFO " + format + "\n", args...)
	}
}

func (self * LogLogger) Debug(format string, args ...interface{}) {
	for _, v := range Range(&self.debug) {
		fmt.Fprintf(v.out, v.datetime() + "DEBUG " + format + "\n", args...)
	}
}

func (self * LogLogger) Trace(format string, args ...interface{}) {
	for _, v := range Range(&self.trace) {
		fmt.Fprintf(v.out, v.datetime() + "TRACE " + format + "\n", args...)
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

func GetLogger() (Logger) {
	return std
}

// SetOutput to std logger
func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

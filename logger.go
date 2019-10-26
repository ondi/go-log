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

type writer_t struct {
	datetime func() string
	out io.Writer
}

type writer_map_t map[string]writer_t

func add_output(value * unsafe.Pointer, name string, out writer_t) {
	for {
		temp := writer_map_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(* writer_map_t)(p) {
			temp[k] = v
		}
		temp[name] = out
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&temp)) {
			return
		}
	}
}

func del_output(value * unsafe.Pointer, name string) {
	for {
		temp := writer_map_t{}
		p := atomic.LoadPointer(value)
		for k, v := range *(* writer_map_t)(p) {
			temp[k] = v
		}
		delete(temp, name)
		if atomic.CompareAndSwapPointer(value, p, unsafe.Pointer(&temp)) {
			return
		}
	}
}

type log_t struct {
	err unsafe.Pointer
	warn unsafe.Pointer
	info unsafe.Pointer
	debug unsafe.Pointer
	trace unsafe.Pointer
}

func NewEmpty() Logger {
	self := &log_t{}
	self.Clear()
	return self
}

func NewLogger(name string, level int, out io.Writer, datetime string) Logger {
	self := NewEmpty()
	self.AddOutput(name, level, out, datetime)
	return self
}

func (self * log_t) AddOutput(name string, level int, out io.Writer, datetime string) {
	writer := writer_t{out: out}
	if len(datetime) > 0 {
		datetime += " "
		writer.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		writer.datetime = func() string {return ""}
	}
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

func (self * log_t) DelOutput(name string) {
	del_output(&self.err, name)
	del_output(&self.warn, name)
	del_output(&self.info, name)
	del_output(&self.debug, name)
	del_output(&self.trace, name)
}

func (self * log_t) Clear() {
	atomic.StorePointer(&self.err, unsafe.Pointer(&writer_map_t{}))
	atomic.StorePointer(&self.warn, unsafe.Pointer(&writer_map_t{}))
	atomic.StorePointer(&self.info, unsafe.Pointer(&writer_map_t{}))
	atomic.StorePointer(&self.debug, unsafe.Pointer(&writer_map_t{}))
	atomic.StorePointer(&self.trace, unsafe.Pointer(&writer_map_t{}))
}

func (self * log_t) Error(format string, args ...interface{}) {
	for _, v := range *(* writer_map_t)(atomic.LoadPointer(&self.err)) {
		fmt.Fprintf(v.out, v.datetime() + "ERROR " + format + "\n", args...)
	}
}

func (self * log_t) Warn(format string, args ...interface{}) {
	for _, v := range *(* writer_map_t)(atomic.LoadPointer(&self.warn)) {
		fmt.Fprintf(v.out, v.datetime() + "WARN " + format + "\n", args...)
	}
}

func (self * log_t) Info(format string, args ...interface{}) {
	for _, v := range *(* writer_map_t)(atomic.LoadPointer(&self.info)) {
		fmt.Fprintf(v.out, v.datetime() + "INFO " + format + "\n", args...)
	}
}

func (self * log_t) Debug(format string, args ...interface{}) {
	for _, v := range *(* writer_map_t)(atomic.LoadPointer(&self.debug)) {
		fmt.Fprintf(v.out, v.datetime() + "DEBUG " + format + "\n", args...)
	}
}

func (self * log_t) Trace(format string, args ...interface{}) {
	for _, v := range *(* writer_map_t)(atomic.LoadPointer(&self.trace)) {
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

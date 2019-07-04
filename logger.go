//
// Log with levels
//

package log

import "os"
import "io"
import "fmt"
import "log"
import "time"
import "sync"
import "sync/atomic"

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
	mx sync.Mutex
	stream io.Writer
	datetime func() string
}

func (self * out_t) Write(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	fmt.Fprintf(self.stream, format, args...)
}

type output_t map[string]*out_t

func add_output(in output_t, name string, value * out_t) (out output_t) {
	out = output_t{}
	for k, v := range in {
		out[k] = v
	}
	out[name] = value
	return
}

func del_output(in output_t, name string) (out output_t) {
	out = output_t{}
	for k, v := range in {
		out[k] = v
	}
	delete(out, name)
	return
}

type LogLogger struct {
	error atomic.Value
	warn atomic.Value
	info atomic.Value
	debug atomic.Value
	trace atomic.Value
}

func NewLogger(name string, level int, out io.Writer, datetime string) Logger {
	self := &LogLogger{}
	self.Clear()
	self.AddOutput(name, level, out, datetime)
	return self
}

func (self * LogLogger) AddOutput(name string, level int, out io.Writer, datetime string) {
	value := &out_t{stream: out}
	if len(datetime) > 0 {
		datetime += " "
		value.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		value.datetime = func() string {return ""}
	}
	if level <= LOG_ERROR {
		self.error.Store(add_output(self.error.Load().(output_t), name, value))
	}
	if level <= LOG_WARN {
		self.warn.Store(add_output(self.warn.Load().(output_t), name, value))
	}
	if level <= LOG_INFO {
		self.info.Store(add_output(self.info.Load().(output_t), name, value))
	}
	if level <= LOG_DEBUG {
		self.debug.Store(add_output(self.debug.Load().(output_t), name, value))
	}
	if level <= LOG_TRACE {
		self.trace.Store(add_output(self.trace.Load().(output_t), name, value))
	}
}

func (self * LogLogger) DelOutput(name string) {
	self.error.Store(del_output(self.error.Load().(output_t), name))
	self.warn.Store(del_output(self.warn.Load().(output_t), name))
	self.info.Store(del_output(self.info.Load().(output_t), name))
	self.debug.Store(del_output(self.debug.Load().(output_t), name))
	self.trace.Store(del_output(self.trace.Load().(output_t), name))
}

func (self * LogLogger) Clear() {
	self.error.Store(output_t{})
	self.warn.Store(output_t{})
	self.info.Store(output_t{})
	self.debug.Store(output_t{})
	self.trace.Store(output_t{})
}

func (self * LogLogger) Error(format string, args ...interface{}) {
	for _, v := range self.error.Load().(output_t) {
		v.Write(v.datetime() + "ERROR " + format + "\n", args...)
	}
}

func (self * LogLogger) Warn(format string, args ...interface{}) {
	for _, v := range self.warn.Load().(output_t) {
		v.Write(v.datetime() + "WARN " + format + "\n", args...)
	}
}

func (self * LogLogger) Info(format string, args ...interface{}) {
	for _, v := range self.info.Load().(output_t) {
		v.Write(v.datetime() + "INFO " + format + "\n", args...)
	}
}

func (self * LogLogger) Debug(format string, args ...interface{}) {
	for _, v := range self.debug.Load().(output_t) {
		v.Write(v.datetime() + "DEBUG " + format + "\n", args...)
	}
}

func (self * LogLogger) Trace(format string, args ...interface{}) {
	for _, v := range self.trace.Load().(output_t) {
		v.Write(v.datetime() + "TRACE " + format + "\n", args...)
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

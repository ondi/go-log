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
	datetime func() string
	
	mx sync.Mutex
	out io.Writer
}

func (self * out_t) Write(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	fmt.Fprintf(self.out, format, args...)
}

type outmap_t map[string]*out_t

type fanout_t struct {
	value atomic.Value
}

func NewFanout() (self * fanout_t) {
	self = &fanout_t{}
	self.Clear()
	return
}

func (self * fanout_t) AddOutput(name string, value * out_t) {
	out := outmap_t{}
	for k, v := range self.value.Load().(outmap_t) {
		out[k] = v
	}
	out[name] = value
	self.value.Store(out)
}

func (self * fanout_t) DelOutput(name string) {
	out := outmap_t{}
	for k, v := range self.value.Load().(outmap_t) {
		out[k] = v
	}
	delete(out, name)
	self.value.Store(out)
}

func (self * fanout_t) Range() outmap_t {
	return self.value.Load().(outmap_t)
}

func (self * fanout_t) Clear() {
	self.value.Store(outmap_t{})
}

type LogLogger struct {
	error * fanout_t
	warn * fanout_t
	info * fanout_t
	debug * fanout_t
	trace * fanout_t
}

func NewLogger(name string, level int, out io.Writer, datetime string) Logger {
	self := &LogLogger{}
	self.error = NewFanout()
	self.warn = NewFanout()
	self.info = NewFanout()
	self.debug = NewFanout()
	self.trace = NewFanout()
	self.AddOutput(name, level, out, datetime)
	return self
}

func (self * LogLogger) AddOutput(name string, level int, out io.Writer, datetime string) {
	value := &out_t{out: out}
	if len(datetime) > 0 {
		datetime += " "
		value.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		value.datetime = func() string {return ""}
	}
	if level <= LOG_ERROR {
		self.error.AddOutput(name, value)
	}
	if level <= LOG_WARN {
		self.warn.AddOutput(name, value)
	}
	if level <= LOG_INFO {
		self.info.AddOutput(name, value)
	}
	if level <= LOG_DEBUG {
		self.debug.AddOutput(name, value)
	}
	if level <= LOG_TRACE {
		self.trace.AddOutput(name, value)
	}
}

func (self * LogLogger) DelOutput(name string) {
	self.error.DelOutput(name)
	self.warn.DelOutput(name)
	self.info.DelOutput(name)
	self.debug.DelOutput(name)
	self.trace.DelOutput(name)
}

func (self * LogLogger) Clear() {
	self.error.Clear()
	self.warn.Clear()
	self.info.Clear()
	self.debug.Clear()
	self.trace.Clear()
}

func (self * LogLogger) Error(format string, args ...interface{}) {
	for _, v := range self.error.Range() {
		v.Write(v.datetime() + "ERROR " + format + "\n", args...)
	}
}

func (self * LogLogger) Warn(format string, args ...interface{}) {
	for _, v := range self.warn.Range() {
		v.Write(v.datetime() + "WARN " + format + "\n", args...)
	}
}

func (self * LogLogger) Info(format string, args ...interface{}) {
	for _, v := range self.info.Range() {
		v.Write(v.datetime() + "INFO " + format + "\n", args...)
	}
}

func (self * LogLogger) Debug(format string, args ...interface{}) {
	for _, v := range self.debug.Range() {
		v.Write(v.datetime() + "DEBUG " + format + "\n", args...)
	}
}

func (self * LogLogger) Trace(format string, args ...interface{}) {
	for _, v := range self.trace.Range() {
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

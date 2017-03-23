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

type mapped_t struct {
	datetime func() string
	stream io.Writer
}

type write_map_t map[string]mapped_t

type LogLogger struct {
	mx sync.Mutex
	error write_map_t
	warn write_map_t
	info write_map_t
	debug write_map_t
	trace write_map_t
}

func NewLogger(name string, level int, out io.Writer, datetime string) Logger {
	self := &LogLogger{}
	self.Clear()
	if out != nil {
		self.AddOutput(name, level, out, datetime)
	}
	return self
}

func (self * LogLogger) AddOutput(name string, level int, out io.Writer, datetime string) {
	self.mx.Lock()
	defer self.mx.Unlock()
	var mapped mapped_t
	if len(datetime) > 0 {
		datetime += " "
		mapped.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		mapped.datetime = func() string {return ""}
	}
	mapped.stream = out
	if level <= LOG_ERROR {
		self.error[name] = mapped
	}
	if level <= LOG_WARN {
		self.warn[name] = mapped
	}
	if level <= LOG_INFO {
		self.info[name] = mapped
	}
	if level <= LOG_DEBUG {
		self.debug[name] = mapped
	}
	if level <= LOG_TRACE {
		self.trace[name] = mapped
	}
}

func (self * LogLogger) Clear() {
	self.error = write_map_t{}
	self.warn = write_map_t{}
	self.info = write_map_t{}
	self.debug = write_map_t{}
	self.trace = write_map_t{}
}

func (self * LogLogger) DelOutput(name string) {
	self.mx.Lock()
	defer self.mx.Unlock()
	delete(self.error, name)
	delete(self.warn, name)
	delete(self.info, name)
	delete(self.debug, name)
	delete(self.trace, name)
}

func (self * LogLogger) Error(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.error {
		fmt.Fprintf(v.stream, v.datetime() + "ERROR " + format + "\n", args...)
	}
}

func (self * LogLogger) Warn(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.warn {
		fmt.Fprintf(v.stream, v.datetime() + "WARN " + format + "\n", args...)
	}
}

func (self * LogLogger) Info(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.info {
		fmt.Fprintf(v.stream, v.datetime() + "INFO " + format + "\n", args...)
	}
}

func (self * LogLogger) Debug(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.debug {
		fmt.Fprintf(v.stream, v.datetime() + "DEBUG " + format + "\n", args...)
	}
}

func (self * LogLogger) Trace(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.trace {
		fmt.Fprintf(v.stream, v.datetime() + "TRACE " + format + "\n", args...)
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

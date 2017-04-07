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

type value_t struct {
	datetime func() string
	stream io.Writer
}

type write_map_t map[string]value_t

type LogLogger struct {
	mx_maps sync.Mutex
	mx_write sync.Mutex
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
	self.mx_maps.Lock()
	defer self.mx_maps.Unlock()
	value := value_t{stream: out}
	if len(datetime) > 0 {
		datetime += " "
		value.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		value.datetime = func() string {return ""}
	}
	if level <= LOG_ERROR {
		self.error[name] = value
	}
	if level <= LOG_WARN {
		self.warn[name] = value
	}
	if level <= LOG_INFO {
		self.info[name] = value
	}
	if level <= LOG_DEBUG {
		self.debug[name] = value
	}
	if level <= LOG_TRACE {
		self.trace[name] = value
	}
}

func (self * LogLogger) Clear() {
	self.mx_maps.Lock()
	defer self.mx_maps.Unlock()
	self.error = write_map_t{}
	self.warn = write_map_t{}
	self.info = write_map_t{}
	self.debug = write_map_t{}
	self.trace = write_map_t{}
}

func (self * LogLogger) DelOutput(name string) {
	self.mx_maps.Lock()
	defer self.mx_maps.Unlock()
	delete(self.error, name)
	delete(self.warn, name)
	delete(self.info, name)
	delete(self.debug, name)
	delete(self.trace, name)
}

func (self * LogLogger) Values(m write_map_t) (res []value_t) {
	self.mx_maps.Lock()
	defer self.mx_maps.Unlock()
	for _, v := range m {
		res = append(res, v)
	}
	return
}

func (self * LogLogger) Write(out io.Writer, str string) {
	self.mx_write.Lock()
	defer self.mx_write.Unlock()
	fmt.Fprint(out, str)
}

func (self * LogLogger) Error(format string, args ...interface{}) {
	for _, v := range self.Values(self.error) {
		self.Write(v.stream, fmt.Sprintf(v.datetime() + "ERROR " + format + "\n", args...))
	}
}

func (self * LogLogger) Warn(format string, args ...interface{}) {
	for _, v := range self.Values(self.warn) {
		self.Write(v.stream, fmt.Sprintf(v.datetime() + "WARN " + format + "\n", args...))
	}
}

func (self * LogLogger) Info(format string, args ...interface{}) {
	for _, v := range self.Values(self.info) {
		self.Write(v.stream, fmt.Sprintf(v.datetime() + "INFO " + format + "\n", args...))
	}
}

func (self * LogLogger) Debug(format string, args ...interface{}) {
	for _, v := range self.Values(self.debug) {
		self.Write(v.stream, fmt.Sprintf(v.datetime() + "DEBUG " + format + "\n", args...))
	}
}

func (self * LogLogger) Trace(format string, args ...interface{}) {
	for _, v := range self.Values(self.trace) {
		self.Write(v.stream, fmt.Sprintf(v.datetime() + "TRACE " + format + "\n", args...))
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

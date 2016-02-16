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

var std = NewLogger(os.Stderr, LOG_TRACE, DATETIME1)

type Logger interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	
	AddOutput(out io.Writer, level int, datetime string)
	DelOutput(out io.Writer)
}

type write_map_t map[io.Writer]func() string

type LogLogger struct {
	mx sync.Mutex
	error write_map_t
	warn write_map_t
	info write_map_t
	debug write_map_t
	trace write_map_t
}

func NewLogger(out io.Writer, level int, datetime string) Logger {
	self := &LogLogger{}
	self.error = write_map_t{}
	self.warn = write_map_t{}
	self.info = write_map_t{}
	self.debug = write_map_t{}
	self.trace = write_map_t{}
	if out != nil {
		self.AddOutput(out, level, datetime)
	}
	return self
}

func (self * LogLogger) AddOutput(out io.Writer, level int, datetime string) {
	self.mx.Lock()
	defer self.mx.Unlock()
	var format func() string
	if len(datetime) > 0 {
		format = func() string {return time.Now().Format(datetime + " ")}
	} else {
		format = func() string {return ""}
	}
	if level <= LOG_ERROR {
		self.error[out] = format
	}
	if level <= LOG_WARN {
		self.warn[out] = format
	}
	if level <= LOG_INFO {
		self.info[out] = format
	}
	if level <= LOG_DEBUG {
		self.debug[out] = format
	}
	if level == LOG_TRACE {
		self.trace[out] = format
	}
}

func (self * LogLogger) DelOutput(out io.Writer) {
	self.mx.Lock()
	defer self.mx.Unlock()
	delete(self.error, out)
	delete(self.warn, out)
	delete(self.info, out)
	delete(self.debug, out)
	delete(self.trace, out)
}

func (self * LogLogger) Error(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.error {
		fmt.Fprintf(k, v() + "ERROR " + format + "\n", args...)
	}
}

func (self * LogLogger) Warn(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.warn {
		fmt.Fprintf(k, v() + "WARN " + format + "\n", args...)
	}
}

func (self * LogLogger) Info(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.info {
		fmt.Fprintf(k, v() + "INFO " + format + "\n", args...)
	}
}

func (self * LogLogger) Debug(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.debug {
		fmt.Fprintf(k, v() + "DEBUG " + format + "\n", args...)
	}
}

func (self * LogLogger) Trace(format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for k, v := range self.trace {
		fmt.Fprintf(k, v() + "TRACE " + format + "\n", args...)
	}
}

func Trace(format string, v ...interface{}) {
	std.Trace(format, v...)
}

func Debug(format string, v ...interface{}) {
	std.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	std.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	std.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	std.Error(format, v...)
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

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
)

var std = NewLogger(LOG_TRACE)

type Logger interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	
	SetDateTime(format string)
	AddOutput(out io.Writer, level int)
	DelOutput(out io.Writer)
}

type write_map_t map[io.Writer]struct{}

type LogLogger struct {
	mx sync.Mutex
	datetime func() string
	error write_map_t
	warn write_map_t
	info write_map_t
	debug write_map_t
	trace write_map_t
}

func NewLogger(level int) Logger {
	self := &LogLogger{}
	self.error = write_map_t{}
	self.warn = write_map_t{}
	self.info = write_map_t{}
	self.debug = write_map_t{}
	self.trace = write_map_t{}
	// self.SetDateTime("2006-01-02 15:04:05.000")
	self.SetDateTime("2006-01-02 15:04:05")
	self.AddOutput(os.Stderr, level)
	return self
}

func (self * LogLogger) AddOutput(out io.Writer, level int) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if level <= LOG_ERROR {
		self.error[out] = struct{}{}
	}
	if level <= LOG_WARN {
		self.warn[out] = struct{}{}
	}
	if level <= LOG_INFO {
		self.info[out] = struct{}{}
	}
	if level <= LOG_DEBUG {
		self.debug[out] = struct{}{}
	}
	if level == LOG_TRACE {
		self.trace[out] = struct{}{}
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

func (self * LogLogger) SetDateTime(format string) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if len(format) > 0 {
		self.datetime = func() string {return time.Now().Format(format + " ")}
	} else {
		self.datetime = func() string {return ""}
	}
}

func (self * LogLogger) Error(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "ERROR " + format + "\n", v...)
	for k, _ := range self.error {
		fmt.Fprint(k, temp)
	}
}

func (self * LogLogger) Warn(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "WARN " + format + "\n", v...)
	for k, _ := range self.warn {
		fmt.Fprintf(k, temp)
	}
}

func (self * LogLogger) Info(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "INFO " + format + "\n", v...)
	for k, _ := range self.info {
		fmt.Fprintf(k, temp)
	}
}

func (self * LogLogger) Debug(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "DEBUG " + format + "\n", v...)
	for k, _ := range self.debug {
		fmt.Fprintf(k, temp)
	}
}

func (self * LogLogger) Trace(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "TRACE " + format + "\n", v...)
	for k, _ := range self.trace {
		fmt.Fprintf(k, temp)
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

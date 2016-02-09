//
// Log levels
//

package log

import "os"
import "io"
import "fmt"
import "log"
import "time"
import "sync"

const (
	LOG_TRACE = iota
	LOG_DEBUG
	LOG_INFO
	LOG_WARN
	LOG_ERROR
)

var std = NewLogger(LOG_TRACE)

type Logger interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	
	DateTime(format string)
	SetOutput(out io.Writer)
	DelOutput(out io.Writer)
}

type LogLogger struct {
	Logger
	mx sync.Mutex
	datetime func() string
	out map[io.Writer]struct{}
}
type LogError struct {
	LogLogger
}
type LogWarn struct {
	LogError
}
type LogInfo struct {
	LogWarn
}
type LogDebug struct {
	LogInfo
}
type LogTrace struct {
	LogDebug
}

func (* LogLogger) Trace(string, ...interface{}) {}
func (* LogLogger) Debug(string, ...interface{}) {}
func (* LogLogger) Info(string, ...interface{}) {}
func (* LogLogger) Warn(string, ...interface{}) {}
func (* LogLogger) Error(string, ...interface{}) {}

func (self * LogLogger) DateTime(format string) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if len(format) > 0 {
		self.datetime = func() string {return time.Now().Format(format + " ")}
	} else {
		self.datetime = func() string {return ""}
	}
}

func (self * LogLogger) SetOutput(out io.Writer) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.out != nil {
		self.out[out] = struct{}{}
	} else {
		self.out = map[io.Writer]struct{}{out: struct{}{}}
	}
}

func (self * LogLogger) DelOutput(out io.Writer) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.out != nil {
		delete(self.out, out)
	}
}

func (self * LogError) Error(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "ERROR " + format + "\n", v...)
	for k, _ := range self.out {
		fmt.Fprint(k, temp)
	}
}

func (self * LogWarn) Warn(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "WARN " + format + "\n", v...)
	for k, _ := range self.out {
		fmt.Fprintf(k, temp)
	}
}

func (self * LogInfo) Info(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "INFO " + format + "\n", v...)
	for k, _ := range self.out {
		fmt.Fprintf(k, temp)
	}
}

func (self * LogDebug) Debug(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "DEBUG " + format + "\n", v...)
	for k, _ := range self.out {
		fmt.Fprintf(k, temp)
	}
}

func (self * LogTrace) Trace(format string, v ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	temp := fmt.Sprintf(self.datetime() + "TRACE " + format + "\n", v...)
	for k, _ := range self.out {
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

func NewLogger(level int) (self Logger) {
	switch level {
	case LOG_TRACE:
		self = &LogTrace{}
	case LOG_DEBUG:
		self = &LogDebug{}
	case LOG_INFO:
		self = &LogInfo{}
	case LOG_WARN:
		self = &LogWarn{}
	default:
		self = &LogError{}
	}
	// self.DateTime("2006-01-02 15:04:05.000")
	self.DateTime("2006-01-02 15:04:05")
	self.SetOutput(os.Stderr)
	return
}

func SetLogger(logger Logger) {
	std = logger
}

func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

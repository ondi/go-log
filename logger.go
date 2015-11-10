//
// Log levels
//

package log

import "os"
import "io"
import "fmt"
import "log"
import "time"

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
	SetDateTime(format string)
	SetOutput(out io.Writer)
}

type LogLogger struct {
	Logger
	datetime func() string
	out io.Writer
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

func (self * LogLogger) SetDateTime(format string) {
	if len(format) > 0 {
		format += " "
		self.datetime = func() string {return time.Now().Format(format)}
	} else {
		self.datetime = func() string {return ""}
	}
}

func (self * LogLogger) SetOutput(out io.Writer) {
	self.out = out
}

func (self * LogError) Error(format string, v ...interface{}) {
	fmt.Fprintf(self.out, self.datetime() + "ERROR " + format + "\n", v...)
}

func (self * LogWarn) Warn(format string, v ...interface{}) {
	fmt.Fprintf(self.out, self.datetime() + "WARN " + format + "\n", v...)
}

func (self * LogInfo) Info(format string, v ...interface{}) {
	fmt.Fprintf(self.out, self.datetime() + "INFO " + format + "\n", v...)
}

func (self * LogDebug) Debug(format string, v ...interface{}) {
	fmt.Fprintf(self.out, self.datetime() + "DEBUG " + format + "\n", v...)
}

func (self * LogTrace) Trace(format string, v ...interface{}) {
	fmt.Fprintf(self.out, self.datetime() + "TRACE " + format + "\n", v...)
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
	// self.SetDateTime("2006-01-02 15:04:05.000")
	self.SetDateTime("2006-01-02 15:04:05")
	self.SetOutput(os.Stderr)
	return
}

func SetLogger(logger Logger) {
	std = logger
}

func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

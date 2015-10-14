//
// Log with levels
//

package log_go

import "os"
import "io"
import "fmt"
import "log"
import "sync"
import "time"

var std = NewLogger(0)

type Logger interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	SetOutput(out io.Writer)
}

type LogLogger struct {
	Logger
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

func (self * LogLogger) Trace(string, ...interface{}) {}
func (self * LogLogger) Debug(string, ...interface{}) {}
func (self * LogLogger) Info(string, ...interface{}) {}
func (self * LogLogger) Warn(string, ...interface{}) {}
func (self * LogLogger) Error(string, ...interface{}) {}
func (self * LogLogger) SetOutput(out io.Writer) {
	self.out = out
}

func (self * LogError) Error(format string, v ...interface{}) {
	fmt.Fprintf(self.out, time.Now().Format("2006-01-02 15:04:05 ERROR ") + format + "\n", v...)
}

func (self * LogWarn) Warn(format string, v ...interface{}) {
	fmt.Fprintf(self.out, time.Now().Format("2006-01-02 15:04:05 WARN ") + format + "\n", v...)
}

func (self * LogInfo) Info(format string, v ...interface{}) {
	fmt.Fprintf(self.out, time.Now().Format("2006-01-02 15:04:05 INFO ") + format + "\n", v...)
}

func (self * LogDebug) Debug(format string, v ...interface{}) {
	fmt.Fprintf(self.out, time.Now().Format("2006-01-02 15:04:05 DEBUG ") + format + "\n", v...)
}

func (self * LogTrace) Trace(format string, v ...interface{}) {
	fmt.Fprintf(self.out, time.Now().Format("2006-01-02 15:04:05 TRACE ") + format + "\n", v...)
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

func NewLogger(level int) (log Logger) {
	switch level {
	case 0:
		log = &LogTrace{}
	case 1:
		log = &LogDebug{}
	case 2:
		log = &LogInfo{}
	case 3:
		log = &LogWarn{}
	default:
		log = &LogError{}
	}
	log.SetOutput(os.Stderr)
	return
}

func SetLogger(logger Logger) {
	std = logger
}

func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

type NoStderr_t struct {}
type Stderr_t struct {}

func (* NoStderr_t) Write(m []byte) (int, error) {
	return 0, nil
}

func (* Stderr_t) Write(m []byte) (int, error) {
	return os.Stderr.Write(m)
}

type RotateLogWriter struct {
	mx sync.Mutex
	fp * os.File
	filename string
	max_bytes int
	curr_bytes int
	backup_count int
	stderr io.Writer
}

func (self * RotateLogWriter) Write(m []byte) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.curr_bytes + len(m) >= self.max_bytes {
		self.LogCycle()
	}
	n, err = self.fp.Write(m)
	self.stderr.Write(m)
	self.curr_bytes += n
	return
}

func (self * RotateLogWriter) LogCycle() {
	if self.fp != nil {
		self.fp.Close()
	}
	os.Remove(fmt.Sprintf("%s.%d", self.filename, self.backup_count))
	for i := self.backup_count; i > 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", self.filename, i - 1), fmt.Sprintf("%s.%d", self.filename, i))
	}
	os.Rename(self.filename, fmt.Sprintf("%s.%d", self.filename, 1))
	self.curr_bytes = 0
	var err error
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY | os.O_CREATE /*| os.O_APPEND*/, 0644)
	if err != nil {
		self.stderr = &Stderr_t{}
		fmt.Fprintln(self.stderr, err.Error())
	}
}

func NewRotateLogWriter(filename string, max_bytes int, backup_count int, stderr bool) (log * RotateLogWriter) {
	var duplicate io.Writer
	if stderr {
		duplicate = &Stderr_t{}
	} else {
		duplicate = &NoStderr_t{}
	}
	log = &RotateLogWriter{filename: filename, max_bytes: max_bytes, backup_count: backup_count, stderr: duplicate}
	log.LogCycle()
	return
}

//
//
//

package log

import "os"
import "fmt"
import "time"

type log_stderr_t struct {
	datetime func() string
}

func NewLogStderr(datetime string) Writer {
	self := &log_stderr_t{}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		self.datetime = func() string {return ""}
	}
	return self
}

func (self * log_stderr_t) Write(level string, format string, args ...interface{}) (err error) {
	_, err = fmt.Fprintf(os.Stderr, self.datetime() + level + " " + format + "\n", args...)
	return
}

type log_stdout_t struct {
	datetime func() string
}

func NewLogStdout(datetime string) Writer {
	self := &log_stdout_t{}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		self.datetime = func() string {return ""}
	}
	return self
}

func (self * log_stdout_t) Write(level string, format string, args ...interface{}) (err error) {
	_, err = fmt.Fprintf(os.Stdout, self.datetime() + level + " " + format + "\n", args...)
	return
}

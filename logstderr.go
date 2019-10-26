//
//
//

package log

import "os"
import "fmt"
import "time"

type stderr_t struct {
	datetime func() string
}

func NewStderr(datetime string) Writer {
	self := &stderr_t{}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		self.datetime = func() string {return ""}
	}
	return self
}

func (self * stderr_t) Write(level string, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, self.datetime() + level + " " + format + "\n", args...)
}

type stdout_t struct {
	datetime func() string
}

func NewStdout(datetime string) Writer {
	self := &stdout_t{}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		self.datetime = func() string {return ""}
	}
	return self
}

func (self * stdout_t) Write(level string, format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, self.datetime() + level + " " + format + "\n", args...)
}

//
//
//

package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Stdany_t struct {
	out      io.Writer
	datetime func() string
}

func NewStdany(out io.Writer, datetime string) Writer {
	self := &Stdany_t{out: out}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string { return time.Now().Format(datetime) }
	} else {
		self.datetime = func() string { return "" }
	}
	return self
}

func (self *Stdany_t) WriteLevel(level string, format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(self.out, self.datetime()+level+" "+format+"\n", args...)
}

func NewStderr(datetime string) Writer {
	return NewStdany(os.Stderr, datetime)
}

func NewStdout(datetime string) Writer {
	return NewStdany(os.Stdout, datetime)
}

//
//
//

package log

import (
	"fmt"
	"io"
	"os"
)

type Stdany_t struct {
	out      io.Writer
	datetime DateTime_t
}

func NewStdany(out io.Writer, datetime DateTime_t) Writer {
	return &Stdany_t{
		out:      out,
		datetime: datetime,
	}
}

func (self *Stdany_t) WriteLevel(level string, format string, args ...interface{}) (int, error) {
	dt := self.datetime()
	if len(dt) > 0 {
		dt += " "
	}
	return fmt.Fprintf(self.out, dt+level+" "+format+"\n", args...)
}

func NewStderr(datetime DateTime_t) Writer {
	return NewStdany(os.Stderr, datetime)
}

func NewStdout(datetime DateTime_t) Writer {
	return NewStdany(os.Stdout, datetime)
}

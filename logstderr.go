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
	prefix []Prefixer
	out    io.Writer
}

func NewStdany(prefix []Prefixer, out io.Writer) Writer {
	return &Stdany_t{
		prefix: prefix,
		out:    out,
	}
}

func (self *Stdany_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	for i, v := range self.prefix {
		if i > 0 {
			io.WriteString(self.out, " ")
		}
		io.WriteString(self.out, v.Prefix())
	}
	if len(self.prefix) > 0 {
		io.WriteString(self.out, " ")
	}
	io.WriteString(self.out, level)
	io.WriteString(self.out, " ")
	n, err = fmt.Fprintf(self.out, format, args...)
	io.WriteString(self.out, "\n")
	return
}

func (self *Stdany_t) Close() error {
	return nil
}

func NewStderr(prefix []Prefixer) Writer {
	return NewStdany(prefix, os.Stderr)
}

func NewStdout(prefix []Prefixer) Writer {
	return NewStdany(prefix, os.Stdout)
}

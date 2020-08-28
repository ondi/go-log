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
	out    io.Writer
	prefix Prefix
}

func NewStdany(out io.Writer, prefix Prefix) Writer {
	return &Stdany_t{
		out:    out,
		prefix: prefix,
	}
}

func (self *Stdany_t) WriteLevel(level string, format string, args ...interface{}) (int, error) {
	p := self.prefix.Prefix()
	if len(p) > 0 {
		p += " "
	}
	return fmt.Fprintf(self.out, p+level+" "+format+"\n", args...)
}

func NewStderr(prefix Prefix) Writer {
	return NewStdany(os.Stderr, prefix)
}

func NewStdout(prefix Prefix) Writer {
	return NewStdany(os.Stdout, prefix)
}

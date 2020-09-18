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
	prefix Prefixer
	out    io.Writer
}

func NewStdany(prefix Prefixer, out io.Writer) Writer {
	return &Stdany_t{
		prefix: prefix,
		out:    out,
	}
}

func (self *Stdany_t) WriteLevel(level string, format string, args ...interface{}) (int, error) {
	p := self.prefix.Prefix()
	if len(p) > 0 {
		p += " "
	}
	return fmt.Fprintf(self.out, p+level+" "+format+"\n", args...)
}

func NewStderr(prefix Prefixer) Writer {
	return NewStdany(prefix, os.Stderr)
}

func NewStdout(prefix Prefixer) Writer {
	return NewStdany(prefix, os.Stdout)
}

//
//
//

package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Stdany_t struct {
	mx     sync.Mutex
	prefix []Formatter
	out    io.Writer
}

func NewStdany(prefix []Formatter, out io.Writer) Queue {
	return &Stdany_t{
		prefix: prefix,
		out:    out,
	}

}

func (self *Stdany_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.prefix {
		v.FormatLog(m.ctx, self.out, m.ts, m.level, m.format, m.args...)
	}
	io.WriteString(self.out, m.level)
	io.WriteString(self.out, " ")
	n, err = fmt.Fprintf(self.out, m.format, m.args...)
	io.WriteString(self.out, "\n")
	return
}

func (self *Stdany_t) ReadLog(count int) (out []Msg_t, oki int) {
	return
}

func (self *Stdany_t) Close() error {
	return nil
}

func NewStderr(prefix []Formatter) Queue {
	return NewStdany(prefix, os.Stderr)
}

func NewStdout(prefix []Formatter) Queue {
	return NewStdany(prefix, os.Stdout)
}

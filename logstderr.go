//
//
//

package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Stdany_t struct {
	mx     sync.Mutex
	prefix []Formatter
	out    io.Writer
}

func NewStdany(prefix []Formatter, out io.Writer) Writer {
	return &Stdany_t{
		prefix: prefix,
		out:    out,
	}
}

func (self *Stdany_t) WriteLog(ctx context.Context, ts time.Time, level string, format string, args ...any) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.prefix {
		v.FormatLog(ctx, self.out, ts, level, format, args...)
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

func NewStderr(prefix []Formatter) Writer {
	return NewStdany(prefix, os.Stderr)
}

func NewStdout(prefix []Formatter) Writer {
	return NewStdany(prefix, os.Stdout)
}

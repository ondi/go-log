//
//
//

package log

import (
	"context"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type Prefixer interface {
	Prefix(ctx context.Context, out io.Writer, ts time.Time, level string, format string) (int, error)
}

type DT_t struct {
	Format string
}

func (self *DT_t) Prefix(ctx context.Context, out io.Writer, ts time.Time, level string, format string) (n int, err error) {
	var b [64]byte
	if n, err = out.Write(ts.AppendFormat(b[:0], self.Format)); n > 0 {
		io.WriteString(out, " ")
	}
	return
}

type FL_t struct{}

func (self *FL_t) Prefix(ctx context.Context, out io.Writer, ts time.Time, level string, format string) (n int, err error) {
	var next_line int
	var next_path string
	_, prev_path, prev_line, ok := runtime.Caller(1)
	for i := 2; i < 100; i++ {
		if _, next_path, next_line, ok = runtime.Caller(i); ok {
			if filepath.Dir(prev_path) != filepath.Dir(next_path) {
				prev_path, prev_line = next_path, next_line
				break
			}
		} else {
			break
		}
	}
	if n, err = io.WriteString(out, filepath.Base(prev_path)); n > 0 {
		io.WriteString(out, ":")
		io.WriteString(out, strconv.FormatInt(int64(prev_line), 10))
		io.WriteString(out, " ")
	}
	return
}

type CX_t struct{}

func (self *CX_t) Prefix(ctx context.Context, out io.Writer, ts time.Time, level string, format string) (n int, err error) {
	if v, _ := ctx.Value(errors_context).(ErrorsContext); v != nil {
		v.Set(level, format)
		if n, err = io.WriteString(out, v.Name()); n > 0 {
			io.WriteString(out, " ")
		}
	}
	return
}

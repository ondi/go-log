//
//
//

package log

import (
	"io"
	"path/filepath"
	"runtime"
	"strconv"
)

func FileLine(skip int, count int) (path string, line int) {
	var next_line int
	var next_path string
	_, path, line, ok := runtime.Caller(skip)
	for i := skip + 1; i < count; i++ {
		if _, next_path, next_line, ok = runtime.Caller(i); !ok {
			return
		}
		if filepath.Dir(path) != filepath.Dir(next_path) {
			return next_path, next_line
		}
	}
	return
}

type DT_t struct {
	Layout string
}

func NewDt(layout string) Formatter {
	return &DT_t{Layout: layout}
}

func (self *DT_t) FormatLog(out io.Writer, m Msg_t) (n int, err error) {
	var b [64]byte
	if n, err = out.Write(m.Level.Ts.AppendFormat(b[:0], self.Layout)); n > 0 {
		io.WriteString(out, " ")
	}
	return
}

type FL_t struct{}

func NewFl() Formatter {
	return &FL_t{}
}

func (self *FL_t) FormatLog(out io.Writer, m Msg_t) (n int, err error) {
	if n, err = io.WriteString(out, filepath.Base(m.Level.File)); n > 0 {
		io.WriteString(out, ":")
		io.WriteString(out, strconv.FormatInt(int64(m.Level.Line), 10))
		io.WriteString(out, " ")
	}
	return
}

type CX_t struct{}

func NewCx() Formatter {
	return &CX_t{}
}

func (self *CX_t) FormatLog(out io.Writer, m Msg_t) (n int, err error) {
	if v := GetErrorsContext(m.Ctx); v != nil {
		v.Set(m.Level, m.Format, m.Args...)
		if n, err = io.WriteString(out, v.Name()); n > 0 {
			io.WriteString(out, " ")
		}
	}
	return
}

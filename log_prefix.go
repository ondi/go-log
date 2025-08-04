//
//
//

package log

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
)

func FileLine(skip int, limit int) (path string, line int) {
	var next_line int
	var next_path string
	_, path, line, ok := runtime.Caller(skip)
	for i := skip + 1; i < limit; i++ {
		if _, next_path, next_line, ok = runtime.Caller(i); !ok {
			return
		}
		if filepath.Dir(path) != filepath.Dir(next_path) {
			return next_path, next_line
		}
	}
	return
}

type PrefixDateTime_t struct {
	Layout string
}

func NewPrefixDateTime(layout string) Formatter {
	return &PrefixDateTime_t{Layout: layout}
}

func (self *PrefixDateTime_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	var b [64]byte
	if n, err = out.Write(in.Info.Ts.AppendFormat(b[:0], self.Layout)); n > 0 {
		io.WriteString(out, " ")
	}
	return
}

type PrefixFileLine_t struct{}

func NewPrefixFileLine() Formatter {
	return &PrefixFileLine_t{}
}

func (self *PrefixFileLine_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	dir, file := filepath.Split(in.Info.File)
	if n, err = io.WriteString(out, filepath.Join(filepath.Base(dir), file)); n > 0 {
		io.WriteString(out, ":")
		io.WriteString(out, strconv.FormatInt(int64(in.Info.Line), 10))
		io.WriteString(out, " ")
	}
	return
}

type PrefixLevelName_t struct {
	prefix string
	suffix string
}

func NewPrefixLevelName(prefix string, suffix string) Formatter {
	return &PrefixLevelName_t{
		prefix: prefix,
		suffix: suffix,
	}
}

func (self *PrefixLevelName_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	io.WriteString(out, self.prefix)
	switch in.Info.Level {
	case 0:
		n, err = io.WriteString(out, "TRACE")
	case 1:
		n, err = io.WriteString(out, "DEBUG")
	case 2:
		n, err = io.WriteString(out, "INFO")
	case 3:
		n, err = io.WriteString(out, "WARN")
	case 4:
		n, err = io.WriteString(out, "ERROR")
	default:
		n, err = fmt.Fprintf(out, "UNDEFINED%v", in.Info.Level)
	}
	io.WriteString(out, self.suffix)
	io.WriteString(out, " ")
	return
}

type PrefixContextName_t struct{}

func NewPrefixContextName() Formatter {
	return &PrefixContextName_t{}
}

func (self *PrefixContextName_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	if v := GetLogContext(in.Ctx); v != nil {
		if n, err = io.WriteString(out, v.ContextName()); n > 0 {
			io.WriteString(out, " ")
		}
	}
	return
}

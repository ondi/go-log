//
//
//

package log

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"
)

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
	n, err = io.WriteString(out, FileLine(in.Info.File, in.Info.Line))
	io.WriteString(out, " ")
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
	n, err = io.WriteString(out, LevelName(in.Info.Level))
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

type PrefixTextMessage_t struct{}

func NewPrefixTextMessage() Formatter {
	return &PrefixTextMessage_t{}
}

func (self *PrefixTextMessage_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	n, err = fmt.Fprintf(out, in.Format, in.Args...)
	return
}

type PrefixNewLine_t struct{}

func NewPrefixNewLine() Formatter {
	return &PrefixNewLine_t{}
}

func (self *PrefixNewLine_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	n, err = fmt.Fprint(out, "\n")
	return
}

type PrefixJsonMessage_t struct {
	AppName     string    `json:"app_name,omitempty"`
	AppVersion  string    `json:"app_version,omitempty"`
	Ts          time.Time `json:"dt,omitempty"`
	Location    string    `json:"location,omitempty"`
	Level       string    `json:"level,omitempty"`
	ContextName string    `json:"context_name,omitempty"`
	Message     string    `json:"message,omitempty"`
}

func NewPrefixJsonMessage(AppName string, AppVersion string) Formatter {
	return &PrefixJsonMessage_t{
		AppName:    AppName,
		AppVersion: AppVersion,
	}
}

func (self *PrefixJsonMessage_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	msg := PrefixJsonMessage_t{
		AppName:    self.AppName,
		AppVersion: self.AppVersion,
		Ts:         in.Info.Ts,
		Location:   FileLine(in.Info.File, in.Info.Line),
		Level:      LevelName(in.Info.Level),
		Message:    fmt.Sprintf(in.Format, in.Args...),
	}
	if v := GetLogContext(in.Ctx); v != nil {
		msg.ContextName = v.ContextName()
	}
	if err = json.NewEncoder(out).Encode(msg); err != nil {
		return
	}
	n++
	return
}

func LevelName(in int64) (res string) {
	switch in {
	case 0:
		res = "TRACE"
	case 1:
		res = "DEBUG"
	case 2:
		res = "INFO"
	case 3:
		res = "WARN"
	case 4:
		res = "ERROR"
	default:
		res = fmt.Sprintf("LEVEL%v", in)
	}
	return
}

func FileLine(f string, l int) (res string) {
	dir, file := filepath.Split(f)
	res = filepath.Join(filepath.Base(dir), file) + ":" + strconv.FormatInt(int64(l), 10)
	return
}

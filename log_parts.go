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

type PartDateTime_t struct {
	Layout string
}

func NewPartDateTime(layout string) Formatter {
	return &PartDateTime_t{Layout: layout}
}

func (self *PartDateTime_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	var b [64]byte
	if n, err = out.Write(in.Info.Ts.AppendFormat(b[:0], self.Layout)); n > 0 {
		io.WriteString(out, " ")
	}
	return
}

type PartFileLine_t struct{}

func NewPartFileLine() Formatter {
	return &PartFileLine_t{}
}

func (self *PartFileLine_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	n, err = io.WriteString(out, FileLine(in.Info.File, in.Info.Line))
	io.WriteString(out, " ")
	return
}

type PartLevelName_t struct {
	prefix string
	suffix string
}

func NewPartLevelName(prefix string, suffix string) Formatter {
	return &PartLevelName_t{
		prefix: prefix,
		suffix: suffix,
	}
}

func (self *PartLevelName_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	io.WriteString(out, self.prefix)
	n, err = io.WriteString(out, LevelName(in.Info.Level))
	io.WriteString(out, self.suffix)
	io.WriteString(out, " ")
	return
}

type PartCircularName_t struct{}

func NewPartCircularName() Formatter {
	return &PartCircularName_t{}
}

func (self *PartCircularName_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	if v := GetLogCircular(in.Ctx); v != nil {
		if n, err = io.WriteString(out, v.CircularName()); n > 0 {
			io.WriteString(out, " ")
		}
	}
	return
}

type PartTextMessage_t struct{}

func NewPartTextMessage() Formatter {
	return &PartTextMessage_t{}
}

func (self *PartTextMessage_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	n, err = fmt.Fprintf(out, in.Format, in.Args...)
	return
}

type PartNewLine_t struct{}

func NewPartNewLine() Formatter {
	return &PartNewLine_t{}
}

func (self *PartNewLine_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	n, err = fmt.Fprint(out, "\n")
	return
}

type PartJsonMessage_t struct {
	AppName     string    `json:"app_name,omitempty"`
	AppVersion  string    `json:"app_version,omitempty"`
	Ts          time.Time `json:"dt,omitempty"`
	Location    string    `json:"location,omitempty"`
	Level       string    `json:"level,omitempty"`
	ContextName string    `json:"context_name,omitempty"`
	Message     string    `json:"message,omitempty"`
}

func NewPartJsonMessage(AppName string, AppVersion string) Formatter {
	return &PartJsonMessage_t{
		AppName:    AppName,
		AppVersion: AppVersion,
	}
}

func (self *PartJsonMessage_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	msg := PartJsonMessage_t{
		AppName:    self.AppName,
		AppVersion: self.AppVersion,
		Ts:         in.Info.Ts,
		Location:   FileLine(in.Info.File, in.Info.Line),
		Level:      LevelName(in.Info.Level),
		Message:    fmt.Sprintf(in.Format, in.Args...),
	}
	if v := GetLogCircular(in.Ctx); v != nil {
		msg.ContextName = v.CircularName()
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

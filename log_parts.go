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

type Tag interface {
	TagKey() string
	TagValue() string
}

type Tag_t struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (self Tag_t) TagKey() string {
	return self.Key
}

func (self Tag_t) TagValue() string {
	return self.Value
}

func (self Tag_t) String() string {
	return self.Key + "=" + self.Value
}

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

type PartCircularId_t struct{}

func NewPartCircularName() Formatter {
	return &PartCircularId_t{}
}

func (self *PartCircularId_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	if v := GetLogCircular(in.Ctx); v != nil {
		if n, err = io.WriteString(out, v.CircularGet("id")); n > 0 {
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
	Level      string            `json:"level,omitempty"`
	Message    string            `json:"message,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
	Location   string            `json:"location,omitempty"`
	ContextId  string            `json:"context_id,omitempty"`
	AppName    string            `json:"app_name,omitempty"`
	AppVersion string            `json:"app_version,omitempty"`
	Ts         time.Time         `json:"dt,omitempty"`
}

func NewPartJsonMessage(AppName string, AppVersion string) Formatter {
	return &PartJsonMessage_t{
		AppName:    AppName,
		AppVersion: AppVersion,
	}
}

func (self *PartJsonMessage_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	msg := PartJsonMessage_t{
		Level:      LevelName(in.Info.Level),
		Message:    fmt.Sprintf(in.Format, in.Args...),
		Location:   FileLine(in.Info.File, in.Info.Line),
		AppName:    self.AppName,
		AppVersion: self.AppVersion,
		Ts:         in.Info.Ts,
	}
	msg.Tags = map[string]string{}
	for _, v := range in.Args {
		if temp, ok := v.(Tag); ok {
			msg.Tags[temp.TagKey()] = temp.TagValue()
		}
	}
	if v := GetLogCircular(in.Ctx); v != nil {
		msg.ContextId = v.CircularGet("id")
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

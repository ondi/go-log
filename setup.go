/*
Logs:
  - LogType: "stdout"
    LogLevel: 0
    LogDate: "2006-01-02 15:04:05"

  - LogType: "file"
    LogLevel: 0
    LogDate: "2006-01-02 15:04:05"
    LogFile: "all.log"
    LogSize: 10000000
    LogDuration: "24h"
    LogBackup: 15

  - LogType: "file"
    LogLevel: 3
    LogDate: "2006-01-02 15:04:05"
    LogFile: "warn.log"
    LogSize: 10000000
    LogDuration: "24h"
    LogBackup: 15

	for k, v := range cfg.Kibana {
		log_http := log.NewHttp(
			64,
			v.Writers,
			log.NewUrls(v.Api),
			log.MessageKB_t{
				ApplicationName: v.AppName,
				Environment:     v.EnvName,
				CallDepth:       4,
				Index: log.MessageIndexKB_t{
					Index: log.MessageIndexNameKB_t{
						Format: v.IndexFormat,
					},
				},
			},
			self.client,
			log.PostHeader(headers),
			log.RpsLimit(log.NewRps(time.Second, 100, 1000)),
		)
		log.GetLogger().AddOutput("http"+strconv.FormatInt(k, 10), log_http, log.WhatLevel(k))
	}
	for k, v := range cfg.Telegram {
		log_tg := log.NewHttp(
			64,
			v.Writers,
			log.NewUrls(v.Api),
			log.MessageTG_t{
				ChatID:    v.ChatID,
				Hostname:  self.hostname,
				TextLimit: 1024,
			},
			self.client,
			log.PostHeader(headers),
			log.PostDelay(1500*time.Millisecond),
		)
		log.GetLogger().AddOutput("telegram"+strconv.FormatInt(k, 10), log_tg, log.WhatLevel(k))
	}
*/

package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type context_key_t string

var context_key context_key_t = "log_ctx"

var std = NewLogger("stderr", NewStderr(&DTFL_t{Format: "2006-01-02 15:04:05", Depth: 4}), WhatLevel(0))

type Context interface {
	Name() string
	Store(level string, format string, args ...interface{})
	Values(func(string, int64) bool)
}

type StoreErrors_t struct {
	name   string
	levels string
	mx     sync.Mutex
	errors map[string]int64
}

func StoreErrorsNew(name string, levels string) Context {
	return &StoreErrors_t{
		name:   name,
		levels: levels,
		errors: map[string]int64{},
	}
}

func (self *StoreErrors_t) Name() string {
	return self.name
}

func (self *StoreErrors_t) Store(level string, format string, args ...interface{}) {
	if strings.Contains(self.levels, level) {
		var res string
		if ix := strings.Index(format, " "); ix > 0 {
			res = format[:ix]
		} else {
			res = format
		}
		self.mx.Lock()
		self.errors[res]++
		if len(self.errors) > 16 {
			for k := range self.errors {
				delete(self.errors, k)
				break
			}
		}
		self.mx.Unlock()
	}
}

func (self *StoreErrors_t) Values(f func(string, int64) bool) {
	self.mx.Lock()
	for k, v := range self.errors {
		if f(k, v) == false {
			break
		}
	}
	self.mx.Unlock()
}

func SetStoreErrorsContext(ctx context.Context, name string, levels string) context.Context {
	return ContextSet(ctx, StoreErrorsNew(name, levels))
}

func ContextSet(ctx context.Context, value Context) context.Context {
	return context.WithValue(ctx, context_key, value)
}

func ContextGet(ctx context.Context) (value Context) {
	value, _ = ctx.Value(context_key).(Context)
	return
}

func ErrorsGet(ctx context.Context, sb *strings.Builder) *strings.Builder {
	v := ContextGet(ctx)
	if v == nil {
		return sb
	}
	var count int
	v.Values(func(in string, num int64) bool {
		if count > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(in)
		sb.WriteString(":")
		sb.WriteString(strconv.FormatInt(num, 10))
		count++
		return true
	})
	return sb
}

func ContextStore(ctx context.Context, level string, format string, args ...interface{}) string {
	if v := ContextGet(ctx); v != nil {
		v.Store(level, format, args...)
		return level + " " + v.Name()
	}
	return level
}

type DT_t struct {
	Format string
}

func (self *DT_t) Prefix() string {
	var b [64]byte
	return string(time.Now().AppendFormat(b[:0], self.Format))
}

type DTFL_t struct {
	Format string
	Depth  int
}

func (self *DTFL_t) Prefix() string {
	var b [64]byte
	_, file, line, ok := runtime.Caller(self.Depth)
	if ok {
		return string(time.Now().AppendFormat(b[:0], self.Format)) + " " + path.Base(file) + ":" + strconv.FormatInt(int64(line), 10)
	}
	return string(time.Now().AppendFormat(b[:0], self.Format))
}

type NoWriter_t struct{}

func (NoWriter_t) WriteLevel(level string, format string, args ...interface{}) (int, error) {
	return 0, nil
}

func (NoWriter_t) Close() error {
	return nil
}

type Args_t struct {
	LogType     string        `yaml:"LogType"`
	LogLevel    int64         `yaml:"LogLevel"`
	LogDate     string        `yaml:"LogDate"`
	LogFile     string        `yaml:"LogFile"`
	LogSize     int           `yaml:"LogSize"`
	LogBackup   int           `yaml:"LogBackup"`
	LogDuration time.Duration `yaml:"LogDuration"`
}

func WhatLevel(in int64) []level_t {
	switch in {
	case 4:
		return LOG_ERROR.Levels
	case 3:
		return LOG_WARN.Levels
	case 2:
		return LOG_INFO.Levels
	case 1:
		return LOG_DEBUG.Levels
	default:
		return LOG_TRACE.Levels
	}
}

func SetupLogger(logs []Args_t) (err error) {
	logger := NewEmpty()
	SetLogger(logger)
	for _, v := range logs {
		switch v.LogType {
		case "file":
			if output, err := NewFileBytes(v.LogFile, &DTFL_t{Format: v.LogDate, Depth: 4}, v.LogSize, v.LogBackup); err != nil {
				Error("LOG FILE: %v", err)
			} else {
				logger.AddOutput(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "filetime":
			if output, err := NewFileTime(v.LogFile, &DTFL_t{Format: v.LogDate, Depth: 4}, v.LogDuration, v.LogBackup); err != nil {
				Error("LOG FILETIME: %v", err)
			} else {
				logger.AddOutput(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "stdout":
			logger.AddOutput("stdout", NewStdout(&DTFL_t{Format: v.LogDate, Depth: 4}), WhatLevel(v.LogLevel))
		case "stderr":
			logger.AddOutput("stderr", NewStderr(&DTFL_t{Format: v.LogDate, Depth: 4}), WhatLevel(v.LogLevel))
		}
	}
	for _, v := range logs {
		Debug("LOG OUTPUT: %v %v %v %v %v", v.LogLevel, v.LogType, v.LogFile, v.LogSize, v.LogBackup)
	}
	return
}

type MessageIndexNameKB_t struct {
	Format string `json:"-"`
	Index  string `json:"_index,omitempty"`
	Type   string `json:"_type,omitempty"`
}

// {"index":{"_index":"logs-2022-01","_type":"_doc"}}
type MessageIndexKB_t struct {
	Index MessageIndexNameKB_t `json:"index"`
}

type MessageKB_t struct {
	ApplicationName string          `json:"ApplicationName"`
	Environment     string          `json:"Environment"`
	Level           string          `json:"Level"`
	Data            json.RawMessage `json:"Data,omitempty"`
	Message         json.RawMessage `json:"Message,omitempty"`
	Timestamp       string          `json:"timestamp"` // "2022-02-12T10:11:52.1862628+03:00"

	// when CallDepth > 0 then Location -> "file:line" from runtime.Caller(CallDepth)
	CallDepth int    `json:"-"`
	Location  string `json:"Location,omitempty"`

	Index MessageIndexKB_t `json:"-"`
}

func (self MessageKB_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
	var b [64]byte
	ts := time.Now()

	if len(self.Index.Index.Format) > 0 {
		self.Index.Index.Index = string(ts.AppendFormat(b[:0], self.Index.Index.Format))
		json.NewEncoder(out).Encode(self.Index)
	}

	self.Level = level
	if strings.HasPrefix(format, "json1") && len(args) > 0 {
		if self.Data, err = json.Marshal(args[0]); err != nil {
			return
		}
	} else if strings.HasPrefix(format, "json") {
		if self.Data, err = json.Marshal(args); err != nil {
			return
		}
	} else {
		if self.Message, err = json.Marshal(level + " " + fmt.Sprintf(format, args...)); err != nil {
			return
		}
	}

	self.Timestamp = string(ts.AppendFormat(b[:0], "2006-01-02T15:04:05.000-07:00"))

	if self.CallDepth > 0 {
		if _, file, line, ok := runtime.Caller(self.CallDepth); ok {
			self.Location = path.Base(file) + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	err = json.NewEncoder(out).Encode(self)
	return
}

type MessageTG_t struct {
	// Unique identifier for the target chat or username of the target channel (in the format @channelusername)
	ChatID int64 `json:"chat_id,omitempty"`
	// Text of the message to be sent
	Text string `json:"text,omitempty"`
	// Optional	Send Markdown or HTML,
	// if you want Telegram apps to show bold, italic,
	// fixed-width text or inline URLs in your bot's message.
	ParseMode string `json:"parse_mode,omitempty"`
	// Optional	Disables link previews for links in this message
	DisableWebPagePreview bool `json:"disable_web_page_preview,omitempty"`
	// Optional	Sends the message silently. Users will receive a notification with no sound.
	DisableNotification bool `json:"disable_notification,omitempty"`
	// Optional	If the message is a reply, ID of the original message
	ReplyToMessageID int64 `json:"reply_to_message_id,omitempty"`
	// Optional	Additional interface options. A JSON-serialized object for an inline keyboard,
	// custom reply keyboard, instructions to remove reply keyboard or to force a reply from the user.
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`

	Hostname  string
	TextLimit int
}

func (self MessageTG_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
	self.Text = self.Hostname + "\n" + level + " " + fmt.Sprintf(format, args...)
	if self.TextLimit > 0 && len(self.Text) > self.TextLimit {
		n := self.TextLimit
		for ; n > 0; n-- {
			if r, _ := utf8.DecodeLastRuneInString(self.Text[:n]); r != utf8.RuneError {
				break
			}
		}
		self.Text = self.Text[:n]
	}
	err = json.NewEncoder(out).Encode(self)
	return
}

func ByteUnit(bytes uint64) (float64, string) {
	switch {
	case bytes >= (1 << (10 * 6)):
		return float64(bytes) / (1 << (10 * 6)), "EB"
	case bytes >= (1 << (10 * 5)):
		return float64(bytes) / (1 << (10 * 5)), "PB"
	case bytes >= (1 << (10 * 4)):
		return float64(bytes) / (1 << (10 * 4)), "TB"
	case bytes >= (1 << (10 * 3)):
		return float64(bytes) / (1 << (10 * 3)), "GB"
	case bytes >= (1 << (10 * 2)):
		return float64(bytes) / (1 << (10 * 2)), "MB"
	case bytes >= (1 << (10 * 1)):
		return float64(bytes) / (1 << (10 * 1)), "KB"
	}
	return float64(bytes), "B"
}

func ByteSize(bytes uint64) string {
	a, b := ByteUnit(bytes)
	return fmt.Sprintf("%.2f %s", a, b)
}

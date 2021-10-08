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

	if len(cfg.KibanaAPI) > 0 {
		self.log_http = log.NewHttp(
			64,
			cfg.KibanaWriters,
			log.NewUrls(cfg.KibanaAPI),
			log.MessageKB_t{
				ApplicationName: cfg.KibanaAppName,
				Environment:     cfg.KibanaEnvName,
				CallDepth:       4,
			},
			self.client,
			log.PostHeader(headers),
			log.RpsLimit(log.NewRps(time.Second, 100, 1000)),
		)
		log.GetLogger().AddOutput("http", log.LOG_WARN, self.log_http)
	} else {
		self.log_http = log.NoWriter_t{}
	}
	if len(cfg.TGBotApi) > 0 {
		self.log_tg = log.NewHttp(
			64,
			cfg.TGWriters,
			log.NewUrls(cfg.TGBotApi+cfg.TGBotToken+"/sendMessage"),
			log.MessageTG_t{
				ChatID:    cfg.TGChatID,
				Hostname:  self.hostname,
				TextLimit: 1024,
			},
			self.client,
			log.PostHeader(headers),
			log.PostDelay(1500*time.Millisecond),
		)
		log.GetLogger().AddOutput("telegram", log.LOG_WARN, self.log_tg)
	} else {
		self.log_tg = log.NoWriter_t{}
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

var std = NewLogger("stderr", LOG_TRACE, NewStderr(&DTFL_t{Format: "2006-01-02 15:04:05", Depth: 4}))

type context_key_t string

type Context interface {
	Value(level string, format string, args ...interface{}) (name string)
}

type Context_t struct {
	name string
	mx   sync.Mutex
	logs map[string]int
}

func (self *Context_t) Value(level string, format string, args ...interface{}) (name string) {
	if level == "ERROR" {
		ix := strings.Index(format, " ")
		if ix > 0 {
			self.mx.Lock()
			self.logs[format[:ix]]++
			self.mx.Unlock()
		}
	}
	return self.name
}

func (self *Context_t) Errors() (res []string) {
	self.mx.Lock()
	for k := range self.logs {
		res = append(res, k)
	}
	self.mx.Unlock()
	return
}

func ContextNew(name string) *Context_t {
	return &Context_t{
		name: name,
		logs: map[string]int{},
	}
}

func ContextSet(ctx context.Context, value Context) context.Context {
	return context.WithValue(ctx, context_key_t("log_ctx"), value)
}

func ContextGet(ctx context.Context) (value Context, ok bool) {
	value, ok = ctx.Value(context_key_t("log_ctx")).(Context)
	return
}

func ContextName(ctx context.Context, level string, format string, args ...interface{}) string {
	if value, ok := ContextGet(ctx); ok {
		return level + " " + value.Value(level, format, args...)
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
	LogLevel    int           `yaml:"LogLevel"`
	LogDate     string        `yaml:"LogDate"`
	LogFile     string        `yaml:"LogFile"`
	LogSize     int           `yaml:"LogSize"`
	LogBackup   int           `yaml:"LogBackup"`
	LogDuration time.Duration `yaml:"LogDuration"`
}

func WhatLevel(in int) level_t {
	switch in {
	case 4:
		return LOG_ERROR
	case 3:
		return LOG_WARN
	case 2:
		return LOG_INFO
	case 1:
		return LOG_DEBUG
	default:
		return LOG_TRACE
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
				logger.AddOutput(v.LogFile, WhatLevel(v.LogLevel), output)
			}
		case "filetime":
			if output, err := NewFileTime(v.LogFile, &DTFL_t{Format: v.LogDate, Depth: 4}, v.LogDuration, v.LogBackup); err != nil {
				Error("LOG FILETIME: %v", err)
			} else {
				logger.AddOutput(v.LogFile, WhatLevel(v.LogLevel), output)
			}
		case "stdout":
			logger.AddOutput("stdout", WhatLevel(v.LogLevel), NewStdout(&DTFL_t{Format: v.LogDate, Depth: 4}))
		case "stderr":
			logger.AddOutput("stderr", WhatLevel(v.LogLevel), NewStderr(&DTFL_t{Format: v.LogDate, Depth: 4}))
		case "dupstderr":
			DupStderr(v.LogFile)
		case "dupstdout":
			DupStdout(v.LogFile)
		}
	}
	for _, v := range logs {
		Debug("LOG OUTPUT: %v %v %v %v %v", v.LogLevel, v.LogType, v.LogFile, v.LogSize, v.LogBackup)
	}
	return
}

type MessageKB_t struct {
	ApplicationName string          `json:"ApplicationName"`
	Environment     string          `json:"Environment"`
	Level           string          `json:"Level"`
	Data            json.RawMessage `json:"Data,omitempty"`
	Message         json.RawMessage `json:"Message,omitempty"`

	// if CallDepth > 0 Location -> "file:line" from runtime.Caller(CallDepth)
	CallDepth int    `json:"-"`
	Location  string `json:"Location,omitempty"`
}

func (self MessageKB_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
	self.Level = level
	if len(format) == 0 {
		if self.Data, err = json.Marshal(args); err != nil {
			return
		}
	} else {
		if self.Message, err = json.Marshal(fmt.Sprintf(format, args...)); err != nil {
			return
		}
	}
	if self.CallDepth > 0 {
		if _, file, line, ok := runtime.Caller(self.CallDepth); ok {
			self.Location = fmt.Sprintf("%s:%d", path.Base(file), line)
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
	self.Text = self.Hostname + "\n" + fmt.Sprintf(format, args...)
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

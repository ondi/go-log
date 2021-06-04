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
			log.Message_t{
				ApplicationName: cfg.KibanaAppName,
				Environment:     cfg.KibanaEnvName,
				CallDepth:       4,
			},
			self.client,
			log.PostHeader(headers),
			log.RpsLimit(log.NewRps(time.Second, 50*time.Millisecond, 1000)),
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
			log.SendMessage_t{
				ChatID:   cfg.TGChatID,
				Hostname: self.hostname,
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
	"encoding/json"
	"fmt"
	"io"
	"path"
	"runtime"
	"time"
)

type Args_t struct {
	LogType     string        `yaml:"LogType"`
	LogLevel    int           `yaml:"LogLevel"`
	LogDate     string        `yaml:"LogDate"`
	LogFile     string        `yaml:"LogFile"`
	LogSize     int           `yaml:"LogSize"`
	LogBackup   int           `yaml:"LogBackup"`
	LogDuration time.Duration `yaml:"LogDuration"`
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
				logger.AddOutput(v.LogFile, v.LogLevel, output)
			}
		case "filetime":
			if output, err := NewFileTime(v.LogFile, &DTFL_t{Format: v.LogDate, Depth: 4}, v.LogDuration, v.LogBackup); err != nil {
				Error("LOG FILETIME: %v", err)
			} else {
				logger.AddOutput(v.LogFile, v.LogLevel, output)
			}
		case "stdout":
			logger.AddOutput("stdout", v.LogLevel, NewStdout(&DTFL_t{Format: v.LogDate, Depth: 4}))
		case "stderr":
			logger.AddOutput("stderr", v.LogLevel, NewStderr(&DTFL_t{Format: v.LogDate, Depth: 4}))
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

	Hostname string
}

func (self MessageTG_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
	self.Text = self.Hostname + "\n" + fmt.Sprintf(format, args...)
	err = json.NewEncoder(out).Encode(self)
	return
}

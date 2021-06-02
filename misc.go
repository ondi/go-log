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
*/

package log

import (
	"encoding/json"
	"fmt"
	"io"
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

// ReplyMarkup	InlineKeyboardMarkup or ReplyKeyboardMarkup or ReplyKeyboardRemove or ForceReply
type SendMessage_t struct {
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

func (self SendMessage_t) Convert(out io.Writer, level string, format string, args ...interface{}) (n int, err error) {
	self.Text = self.Hostname + "\n" + fmt.Sprintf(format, args...)
	err = json.NewEncoder(out).Encode(self)
	return
}

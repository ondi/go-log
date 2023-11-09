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
			log.NewUrls(v.Host),
			log.MessageKB_t{
				ApplicationName: v.AppName,
				Environment:     v.EnvName,
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
		log.GetLogger().AddOutput(k, log_http, log.WhatLevel(v.Level))
	}
	for k, v := range cfg.Telegram {
		log_tg := log.NewHttp(
			64,
			v.Writers,
			log.NewUrls(v.Host),
			log.MessageTG_t{
				ChatID:    v.ChatID,
				Hostname:  self.hostname,
				TextLimit: 1024,
			},
			self.client,
			log.PostHeader(headers),
			log.PostDelay(1500*time.Millisecond),
		)
		log.GetLogger().AddOutput(k, log_tg, log.WhatLevel(v.Level)[:1])
	}
*/

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	STDERR  = os.Stderr
	LEVELS  = []Level_t{LOG_ERROR, LOG_WARN, LOG_INFO, LOG_DEBUG, LOG_TRACE}
	__std   = New(LEVELS).AddOutput("stderr", NewStdany([]Formatter{NewDt("2006-01-02 15:04:05.000"), NewFl(), NewCx()}, os.Stderr), WhatLevel(0))
	__prefs = []Formatter{NewFl(), NewCx()}
)

var (
	LOG_TRACE = Level_t{Name: "TRACE", Level: 0}
	LOG_DEBUG = Level_t{Name: "DEBUG", Level: 1}
	LOG_INFO  = Level_t{Name: "INFO", Level: 2}
	LOG_WARN  = Level_t{Name: "WARN", Level: 3}
	LOG_ERROR = Level_t{Name: "ERROR", Level: 4}
)

type NoWriter_t struct{}

func (NoWriter_t) WriteLog(Msg_t) (int, error) {
	return 0, nil
}

func (NoWriter_t) ReadLog(int) ([]Msg_t, int) {
	return nil, -1
}

func (NoWriter_t) WriteError(int) {
}

func (NoWriter_t) Size() (res QueueSize_t) {
	return
}

func (NoWriter_t) Close() error {
	return nil
}

func NoWriter() Queue {
	return NoWriter_t{}
}

type Args_t struct {
	LogType     string        `yaml:"LogType"`
	LogFile     string        `yaml:"LogFile"`
	LogDate     string        `yaml:"LogDate"`
	LogLevel    int64         `yaml:"LogLevel"`
	LogSize     int           `yaml:"LogSize"`
	LogBackup   int           `yaml:"LogBackup"`
	LogQueue    int           `yaml:"LogQueue"`
	LogWriters  int           `yaml:"LogWriters"`
	LogDuration time.Duration `yaml:"LogDuration"`
}

func WhatLevel(in int64) []Level_t {
	switch in {
	case 4:
		return []Level_t{LOG_ERROR}
	case 3:
		return []Level_t{LOG_ERROR, LOG_WARN}
	case 2:
		return []Level_t{LOG_ERROR, LOG_WARN, LOG_INFO}
	case 1:
		return []Level_t{LOG_ERROR, LOG_WARN, LOG_INFO, LOG_DEBUG}
	default:
		return []Level_t{LOG_ERROR, LOG_WARN, LOG_INFO, LOG_DEBUG, LOG_TRACE}
	}
}

func SetupLogger(ts time.Time, logs []Args_t) (err error) {
	logger := SetLogger(New(LEVELS))
	for _, v := range logs {
		switch v.LogType {
		case "file":
			if output, err := NewFileBytes(ts, v.LogFile, []Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, v.LogSize, v.LogBackup); err != nil {
				fmt.Fprintf(STDERR, "LOG ERROR: %v %v\n", ts.Format("2006-01-02 15:04:05"), err)
			} else {
				logger.AddOutput(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "filequeue":
			if output, err := NewFileBytesQueue(v.LogQueue, v.LogWriters, ts, v.LogFile, []Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, v.LogSize, v.LogBackup); err != nil {
				fmt.Fprintf(STDERR, "LOG ERROR: %v %v\n", ts.Format("2006-01-02 15:04:05"), err)
			} else {
				logger.AddOutput(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "filetime":
			if output, err := NewFileTime(ts, v.LogFile, []Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, v.LogDuration, v.LogBackup); err != nil {
				fmt.Fprintf(STDERR, "LOG ERROR: %v %v\n", ts.Format("2006-01-02 15:04:05"), err)
			} else {
				logger.AddOutput(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "filetimequeue":
			if output, err := NewFileTimeQueue(v.LogQueue, v.LogWriters, ts, v.LogFile, []Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, v.LogDuration, v.LogBackup); err != nil {
				fmt.Fprintf(STDERR, "LOG ERROR: %v %v\n", ts.Format("2006-01-02 15:04:05"), err)
			} else {
				logger.AddOutput(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "stdout":
			logger.AddOutput("stdout", NewStdany([]Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, os.Stdout), WhatLevel(v.LogLevel))
		case "stdoutqueue":
			logger.AddOutput("stdout", NewStdanyQueue(v.LogQueue, v.LogWriters, []Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, os.Stdout), WhatLevel(v.LogLevel))
		case "stderr":
			logger.AddOutput("stderr", NewStdany([]Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, os.Stderr), WhatLevel(v.LogLevel))
		case "stderrqueue":
			logger.AddOutput("stderr", NewStdanyQueue(v.LogQueue, v.LogWriters, []Formatter{NewDt(v.LogDate), NewFl(), NewCx()}, os.Stderr), WhatLevel(v.LogLevel))
		}
	}
	for _, v := range logs {
		Debug("LOG OUTPUT: LogLevel=%v, LogType=%v, LogFile=%v, LogSize=%v, LogDuration=%v, LogBackup=%v, LogQueue=%v, LogWriters=%v",
			v.LogLevel, v.LogType, v.LogFile, ByteSize(uint64(v.LogSize)), v.LogDuration, v.LogBackup, v.LogQueue, v.LogWriters)
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
	Index           MessageIndexKB_t `json:"-"`
	ApplicationName string           `json:"ApplicationName"`
	Environment     string           `json:"Environment"`
	Level           string           `json:"Level"`
	Timestamp       string           `json:"timestamp"` // "2022-02-12T10:11:52.1862628+03:00"
	Location        string           `json:"Location,omitempty"`
	Data            json.RawMessage  `json:"Data,omitempty"`
	Message         json.RawMessage  `json:"Message,omitempty"`
}

func (self MessageKB_t) FormatLog(out io.Writer, m Msg_t) (n int, err error) {
	var b [64]byte

	if len(self.Index.Index.Format) > 0 {
		self.Index.Index.Index = string(m.Level.Ts.AppendFormat(b[:0], self.Index.Index.Format))
		json.NewEncoder(out).Encode(self.Index)
	}

	self.Level = m.Level.Name
	if strings.HasPrefix(m.Format, "json") {
		if self.Data, err = json.Marshal(m.Args); err != nil {
			return
		}
	} else {
		if self.Message, err = json.Marshal(fmt.Sprintf(m.Format, m.Args...)); err != nil {
			return
		}
	}

	self.Timestamp = string(m.Level.Ts.AppendFormat(b[:0], "2006-01-02T15:04:05.000-07:00"))

	var temp bytes.Buffer
	for _, v := range __prefs {
		v.FormatLog(&temp, m)
	}
	self.Location = temp.String()

	err = json.NewEncoder(out).Encode(self)
	return
}

type MessageTG_t struct {
	// Unique identifier for the target chat or username of the target channel (in the format @channelusername)
	ChatId int64 `json:"chat_id,omitempty"`
	// Text of the message to be sent
	Text string `json:"text,omitempty"`

	ApplicationName string `json:"-"`
	Hostname        string `json:"-"`
	TextLimit       int    `json:"-"`
}

func (self MessageTG_t) FormatLog(out io.Writer, m Msg_t) (n int, err error) {
	if len(self.Hostname) > 0 {
		self.Text += self.Hostname + " "
	}

	var w io.Writer
	var buf strings.Builder
	if self.TextLimit > 0 {
		w = &LimitWriter_t{Out: &buf, Limit: self.TextLimit}
	} else {
		w = &buf
	}
	for _, v := range __prefs {
		v.FormatLog(w, m)
	}

	if len(m.Level.Name) > 0 {
		io.WriteString(w, m.Level.Name)
		io.WriteString(w, " ")
	}

	if len(self.ApplicationName) > 0 {
		io.WriteString(w, self.ApplicationName)
		io.WriteString(w, " ")
	}

	fmt.Fprintf(w, m.Format, m.Args...)

	self.Text = buf.String()
	err = json.NewEncoder(out).Encode(self)

	return
}

type LimitWriter_t struct {
	Out     io.Writer
	Limit   int
	written int
}

func (self *LimitWriter_t) Write(p []byte) (n int, err error) {
	if n = self.Limit - self.written; n > len(p) {
		n, err = self.Out.Write(p)
	} else {
		for ; n > 0; n-- {
			if r, _ := utf8.DecodeLastRune(p[:n]); r != utf8.RuneError {
				break
			}
		}
		n, err = self.Out.Write(p[:n])
	}
	self.written += n
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
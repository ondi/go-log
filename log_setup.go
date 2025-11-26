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
		log_http := log.NewWriterHttp(
			log.NewUrls(v.Host),
			log.MessageKB_t{
				ApplicationName: v.AppName,
				Environment:     v.EnvName,
				Hostname:        self.hostname,
				Index: log.MessageIndexKB_t{
					Index: log.MessageIndexNameKB_t{
						Format: v.IndexFormat,
					},
				},
			},
			self.client,
			log.PostHeader(v),
			log.RpsLimit(log.NewRps(time.Second, 100, 1000)),
		)
		log.GetLogger().SwapLevelMap(log.GetLogger().CopyLevelMap().AddOutputs(k, log.NewQueue(1024, v.Writers, 1024, log_http), log.WhatLevel(v.Level)))
		self.log_http[k] = log_http
	}
	for k, v := range cfg.Telegram {
		log_tg := log.NewWriterHttp(
			log.NewUrls(v.Host),
			log.MessageTG_t{
				ChatId:    v.ChatID,
				Hostname:  self.hostname,
				TextLimit: 4096,
			},
			self.client,
			log.PostHeader(v),
			log.PostDelay(1500*time.Millisecond),
		)
		log.GetLogger().SwapLevelMap(log.GetLogger().CopyLevelMap().AddOutputs(k, log.NewQueue(64, v.Writers, 1, log_tg), log.WhatLevel(v.Level)))
		self.log_tg[k] = log_tg
	}
*/

package log

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	__std_logger = NewLogger()
	__std_parts  = []Formatter{NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", "")}
)

type Args_t struct {
	LogType     string        `yaml:"LogType"`
	LogFile     string        `yaml:"LogFile"`
	LogDate     string        `yaml:"LogDate"`
	LogLevel    int64         `yaml:"LogLevel"`
	LogLimit    int           `yaml:"LogLimit"`
	LogSize     int           `yaml:"LogSize"`
	LogBackup   int           `yaml:"LogBackup"`
	LogQueue    int           `yaml:"LogQueue"`
	LogWriters  int           `yaml:"LogWriters"`
	LogDuration time.Duration `yaml:"LogDuration"`
}

func NewLogger() (out Logger) {
	m := NewLevelMap()
	w1 := NewWriterStdany(
		[]Formatter{
			NewPartDateTime("2006-01-02 15:04:05.000"),
			NewPartFileLine(),
			NewPartBufferId(),
			NewPartLevelName("", ""),
			NewPartTextMessage(),
			NewPartNewLine(),
		},
		os.Stderr,
		0,
	)
	w2 := NewLogBufferWriter()
	for _, v := range WhatLevel(0) {
		m.AddOutput(v, "stderr", w1)
		m.AddOutput(v, "buf", w2)
	}
	out = New(m)
	return
}

func WhatLevel(in int64) []int64 {
	switch in {
	case 4:
		return []int64{4}
	case 3:
		return []int64{4, 3}
	case 2:
		return []int64{4, 3, 2}
	case 1:
		return []int64{4, 3, 2, 1}
	default:
		return []int64{4, 3, 2, 1, 0}
	}
}

func LogStderr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	io.WriteString(os.Stderr, "\n")
}

func SetupLogger(ts time.Time, logs []Args_t, app_name string, app_version string) (out Logger, errs []string) {
	m := NewLevelMap()
	for _, v := range logs {
		switch v.LogType {
		case "buf":
			m.AddOutputs("buf", NewLogBufferWriter(), WhatLevel(v.LogLevel))
		case "file":
			if output, err := NewWriterFileBytes(ts, v.LogFile, []Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, v.LogSize, v.LogBackup, v.LogLimit); err != nil {
				errs = append(errs, fmt.Sprintf("%v %v", v.LogType, err.Error()))
			} else {
				m.AddOutputs(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "q_file":
			fq, err := NewWriterFileBytes(ts, v.LogFile, []Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, v.LogSize, v.LogBackup, v.LogLimit)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%v %v", v.LogType, err.Error()))
			} else {
				m.AddOutputs(v.LogFile, NewQueue(v.LogQueue, v.LogWriters, 1, fq), WhatLevel(v.LogLevel))
			}
		case "filetime":
			if output, err := NewWriterFileTime(ts, v.LogFile, []Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, v.LogDuration, v.LogBackup, v.LogLimit); err != nil {
				errs = append(errs, fmt.Sprintf("%v %v", v.LogType, err.Error()))
			} else {
				m.AddOutputs(v.LogFile, output, WhatLevel(v.LogLevel))
			}
		case "q_filetime":
			fq, err := NewWriterFileTime(ts, v.LogFile, []Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, v.LogDuration, v.LogBackup, v.LogLimit)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%v %v", v.LogType, err.Error()))
			} else {
				m.AddOutputs(v.LogFile, NewQueue(v.LogQueue, v.LogWriters, 1, fq), WhatLevel(v.LogLevel))
			}
		case "stdout":
			m.AddOutputs("stdout", NewWriterStdany([]Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, os.Stdout, v.LogLimit), WhatLevel(v.LogLevel))
		case "stdout2":
			m.AddOutputs("stdout2", NewWriterStdany([]Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("_", "_"), NewPartTextMessage(), NewPartNewLine()}, os.Stdout, v.LogLimit), WhatLevel(v.LogLevel))
		case "q_stdout":
			q := NewWriterStdany([]Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, os.Stdout, v.LogLimit)
			m.AddOutputs("stdoutqueue", NewQueue(v.LogQueue, v.LogWriters, 1, q), WhatLevel(v.LogLevel))
		case "stderr":
			m.AddOutputs("stderr", NewWriterStdany([]Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, os.Stderr, v.LogLimit), WhatLevel(v.LogLevel))
		case "q_stderr":
			q := NewWriterStdany([]Formatter{NewPartDateTime(v.LogDate), NewPartFileLine(), NewPartBufferId(), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, os.Stderr, v.LogLimit)
			m.AddOutputs("stderrqueue", NewQueue(v.LogQueue, v.LogWriters, 1, q), WhatLevel(v.LogLevel))
		case "q_json_stdout":
			q := NewWriterStdany([]Formatter{NewPartJsonMessage(app_name, app_version)}, os.Stdout, v.LogLimit)
			m.AddOutputs("stderrqueue", NewQueue(v.LogQueue, v.LogWriters, 1, q), WhatLevel(v.LogLevel))
		case "q_json_stderr":
			q := NewWriterStdany([]Formatter{NewPartJsonMessage(app_name, app_version)}, os.Stderr, v.LogLimit)
			m.AddOutputs("stderrqueue", NewQueue(v.LogQueue, v.LogWriters, 1, q), WhatLevel(v.LogLevel))
		}
	}
	out = New(m)
	return
}

func SetupPrint(logs []Args_t, errs []string, log_debug func(string, ...any)) {
	for _, v := range logs {
		log_debug("LOG OUTPUT: LogLevel=%v, LogLimit=%v, LogType=%v, LogFile=%v, LogSize=%v, LogDuration=%v, LogBackup=%v, LogQueue=%v, LogWriters=%v",
			v.LogLevel, v.LogLimit, v.LogType, v.LogFile, ByteSize(uint64(v.LogSize)), v.LogDuration, v.LogBackup, v.LogQueue, v.LogWriters)
	}
	log_debug("LOG SETUP ERRORS: %v", errs)
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
	Timestamp       string           `json:"timestamp"` // 2006-01-02T15:04:05.000-07:00
	ApplicationName string           `json:"ApplicationName"`
	Environment     string           `json:"Environment"`
	Level           string           `json:"Level"`
	Location        string           `json:"Location,omitempty"`
	Hostname        string           `json:"Hostname,omitempty"`
	Message         string           `json:"Message,omitempty"`
	Data            json.RawMessage  `json:"Data,omitempty"`
	TextLimit       int              `json:"-"`
}

func (self MessageKB_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	var b [64]byte
	var buf strings.Builder

	if self.TextLimit == 0 {
		self.TextLimit = math.MaxInt
	}

	w := &LimitWriter_t{Buf: &buf, Limit: self.TextLimit}

	buf.Reset()
	w.Limit = self.TextLimit

	if len(self.Index.Index.Format) > 0 {
		self.Index.Index.Index = string(in.Info.Ts.AppendFormat(b[:0], self.Index.Index.Format))
		json.NewEncoder(out).Encode(self.Index)
	}

	if strings.HasPrefix(in.Format, "json") {
		if self.Data, err = json.Marshal(in.Args); err != nil {
			return
		}
	} else {
		fmt.Fprintf(w, in.Format, in.Args...)
		self.Message = buf.String()
	}

	self.Level = fmt.Sprintf("LEVEL%v", in.Info.Level)
	self.Timestamp = string(in.Info.Ts.AppendFormat(b[:0], "2006-01-02T15:04:05.000-07:00"))

	buf.Reset()
	for _, fm := range __std_parts {
		fm.FormatMessage(&buf, in)
	}
	self.Location = buf.String()

	if err = json.NewEncoder(out).Encode(self); err != nil {
		return
	}
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

func (self MessageTG_t) FormatMessage(out io.Writer, in Msg_t) (n int, err error) {
	var buf strings.Builder

	if self.TextLimit == 0 {
		self.TextLimit = math.MaxInt
	}

	w := &LimitWriter_t{Buf: &buf, Limit: self.TextLimit}

	if len(self.Hostname) > 0 {
		io.WriteString(w, self.Hostname)
		io.WriteString(w, " ")
	}

	if len(self.ApplicationName) > 0 {
		io.WriteString(w, self.ApplicationName)
		io.WriteString(w, " ")
	}

	for _, fm := range __std_parts {
		fm.FormatMessage(w, in)
	}

	fmt.Fprintf(w, in.Format, in.Args...)
	fmt.Fprintf(w, "\n")

	self.Text = buf.String()
	if err = json.NewEncoder(out).Encode(self); err != nil {
		return
	}
	return
}

type LimitWriter_t struct {
	Buf   io.Writer
	Limit int
}

func (self *LimitWriter_t) Write(p []byte) (n int, err error) {
	if self.Limit >= len(p) {
		n, err = self.Buf.Write(p)
	} else {
		for ; self.Limit > 0; self.Limit-- {
			if r, _ := utf8.DecodeLastRune(p[:self.Limit]); r != utf8.RuneError {
				break
			}
		}
		n, err = self.Buf.Write(p[:self.Limit])
	}
	self.Limit -= n
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

type TransportOption func(self *http.Transport)

func DialContext(timeout time.Duration) TransportOption {
	return func(self *http.Transport) {
		self.DialContext = (&net.Dialer{Timeout: timeout}).DialContext
	}
}

func MaxIdleConns(n int) TransportOption {
	return func(self *http.Transport) {
		self.MaxIdleConns = n
	}
}

func MaxIdleConnsPerHost(n int) TransportOption {
	return func(self *http.Transport) {
		self.MaxIdleConnsPerHost = n
	}
}

func ProxyFromEnvironment() TransportOption {
	return func(self *http.Transport) {
		self.Proxy = http.ProxyFromEnvironment
	}
}

func DisableCompression() TransportOption {
	return func(self *http.Transport) {
		self.DisableCompression = true
	}
}

func ForceAttemptHTTP2() TransportOption {
	return func(self *http.Transport) {
		self.ForceAttemptHTTP2 = true
	}
}

func InsecureSkipVerify() TransportOption {
	return func(self *http.Transport) {
		self.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

func NewTransport(opts ...TransportOption) http.RoundTripper {
	tr := &http.Transport{
		// do not set IdleConnTimeout to zero
		// (pprof) top 30
		// Showing nodes accounting for 988.20MB, 93.70% of 1054.61MB total
		// Dropped 253 nodes (cum <= 5.27MB)
		// Showing top 30 nodes out of 55
		//       flat  flat%   sum%        cum   cum%
		//   850.04MB 80.60% 80.60%   853.54MB 80.93%  net.(*Resolver).exchange
		IdleConnTimeout: 90 * time.Second,
	}
	for _, v := range opts {
		v(tr)
	}
	return tr
}

// Default
// MaxIdleConns:        100,
// MaxIdleConnsPerHost: 2,
func DefaultTransport(dial_timeout time.Duration, MaxIdleConns int, MaxIdleConnsPerHost int) http.RoundTripper {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: dial_timeout}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		// do not set IdleConnTimeout to zero
		// (pprof) top 30
		// Showing nodes accounting for 988.20MB, 93.70% of 1054.61MB total
		// Dropped 253 nodes (cum <= 5.27MB)
		// Showing top 30 nodes out of 55
		//       flat  flat%   sum%        cum   cum%
		//   850.04MB 80.60% 80.60%   853.54MB 80.93%  net.(*Resolver).exchange
		IdleConnTimeout: 90 * time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

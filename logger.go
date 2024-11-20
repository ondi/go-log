/*
	Log(levels) with no allocation and locks
*/

package log

import (
	"context"
	"io"
	"sync/atomic"
	"time"
)

type Info_t struct {
	Ts        time.Time `json:"ts"`
	LevelName string    `json:"level_name"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	LevelId   int64     `json:"level"`
}

func (self *Info_t) Set(ts time.Time) {
	self.Ts = ts
	self.File, self.Line = FileLine(1, 32)
}

type Msg_t struct {
	Ctx    context.Context `json:"-"`
	Info   Info_t          `json:"info"`
	Format string          `json:"format"`
	Args   []any           `json:"args"`
}

type QueueSize_t struct {
	Limit         int
	Size          int
	Readers       int
	Writers       int
	QueueWrite    int
	QueueRead     int
	QueueOverflow int
	WriteErrorCnt int
	WriteErrorMsg string
}

type Queue interface {
	LogWrite(m Msg_t) (int, error)
	Size() QueueSize_t
	Close() error
}

type Formatter interface {
	FormatMessage(out io.Writer, in ...Msg_t) (int, error)
}

type Logger interface {
	Log(ctx context.Context, level Info_t, format string, args ...any)

	Trace(format string, args ...any)
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)

	TraceCtx(ctx context.Context, format string, args ...any)
	DebugCtx(ctx context.Context, format string, args ...any)
	InfoCtx(ctx context.Context, format string, args ...any)
	WarnCtx(ctx context.Context, format string, args ...any)
	ErrorCtx(ctx context.Context, format string, args ...any)

	SwapLevelMap(Level_map_t) Level_map_t
	CopyLevelMap() Level_map_t

	Range(fn func(level_id int64, writer_name string, writer Queue) bool)
}

type log_t struct {
	level_map atomic.Pointer[Level_map_t]
}

// use NewLevelMap()
func New(in Level_map_t) Logger {
	self := &log_t{}
	temp := in.Copy(Level_map_t{})
	self.level_map.Store(&temp)
	return self
}

func (self *log_t) SwapLevelMap(in Level_map_t) Level_map_t {
	temp := in.Copy(Level_map_t{})
	return *self.level_map.Swap(&temp)
}

func (self *log_t) CopyLevelMap() (out Level_map_t) {
	return (*self.level_map.Load()).Copy(Level_map_t{})
}

func (self *log_t) Range(fn func(level_id int64, writer_name string, writer Queue) bool) {
	for level_id, level := range *self.level_map.Load() {
		for writer_name, writer := range level {
			if fn(level_id, writer_name, writer) == false {
				return
			}
		}
	}
}

func (self *log_t) Log(ctx context.Context, level Info_t, format string, args ...any) {
	level.Set(time.Now())
	for _, writer := range (*self.level_map.Load())[level.LevelId] {
		writer.LogWrite(Msg_t{Ctx: ctx, Info: level, Format: format, Args: args})
	}
}

func (self *log_t) Error(format string, args ...any) {
	self.Log(context.Background(), LOG_ERROR, format, args...)
}

func (self *log_t) Warn(format string, args ...any) {
	self.Log(context.Background(), LOG_WARN, format, args...)
}

func (self *log_t) Info(format string, args ...any) {
	self.Log(context.Background(), LOG_INFO, format, args...)
}

func (self *log_t) Debug(format string, args ...any) {
	self.Log(context.Background(), LOG_DEBUG, format, args...)
}

func (self *log_t) Trace(format string, args ...any) {
	self.Log(context.Background(), LOG_TRACE, format, args...)
}

func (self *log_t) ErrorCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_ERROR, format, args...)
}

func (self *log_t) WarnCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_WARN, format, args...)
}

func (self *log_t) InfoCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_INFO, format, args...)
}

func (self *log_t) DebugCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_DEBUG, format, args...)
}

func (self *log_t) TraceCtx(ctx context.Context, format string, args ...any) {
	self.Log(ctx, LOG_TRACE, format, args...)
}

func Error(format string, args ...any) {
	__std.Error(format, args...)
}

func Warn(format string, args ...any) {
	__std.Warn(format, args...)
}

func Info(format string, args ...any) {
	__std.Info(format, args...)
}

func Debug(format string, args ...any) {
	__std.Debug(format, args...)
}

func Trace(format string, args ...any) {
	__std.Trace(format, args...)
}

func ErrorCtx(ctx context.Context, format string, args ...any) {
	__std.ErrorCtx(ctx, format, args...)
}

func WarnCtx(ctx context.Context, format string, args ...any) {
	__std.WarnCtx(ctx, format, args...)
}

func InfoCtx(ctx context.Context, format string, args ...any) {
	__std.InfoCtx(ctx, format, args...)
}

func DebugCtx(ctx context.Context, format string, args ...any) {
	__std.DebugCtx(ctx, format, args...)
}

func TraceCtx(ctx context.Context, format string, args ...any) {
	__std.TraceCtx(ctx, format, args...)
}

func SetLogger(in Logger) Logger {
	__std = in
	return __std
}

func GetLogger() Logger {
	return __std
}

//
//
//

package log

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// &log_ctx used for ctx.Value
var log_ctx = 1

type LogContext interface {
	Name() string
	Set(m LogMsg_t)
	Range(func(level string, format string, args ...any) bool)
	Reset()
}

type LogContext_t struct {
	mx     sync.Mutex
	levels map[int64]Level_t
	name   string
	data   []LogMsg_t
	limit  int
}

func NewLogContext(name string, limit int, levels []Level_t) LogContext {
	self := &LogContext_t{
		name:   name,
		levels: map[int64]Level_t{},
		limit:  limit,
	}
	for _, v := range levels {
		self.levels[v.Level] = v
	}
	return self
}

func (self *LogContext_t) Name() string {
	return self.name
}

func (self *LogContext_t) Set(m LogMsg_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if _, ok := self.levels[m.Level.Level]; ok {
		if len(self.data) < self.limit {
			self.data = append(self.data, m)
		}
	}
}

func (self *LogContext_t) Range(f func(level string, format string, args ...any) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range self.data {
		if f(v.Level.Name, v.Format, v.Args...) == false {
			break
		}
	}
}

func (self *LogContext_t) Reset() {
	self.data = nil
}

func SetLogContextNew(ctx context.Context, name string, limit int, levels []Level_t) context.Context {
	return context.WithValue(ctx, &log_ctx, NewLogContext(name, limit, levels))
}

func SetLogContextValue(ctx context.Context, value LogContext) context.Context {
	return context.WithValue(ctx, &log_ctx, value)
}

func GetLogContextValue(ctx context.Context) (value LogContext) {
	value, _ = ctx.Value(&log_ctx).(LogContext)
	return
}

func GetLogContextPayload(ctx context.Context, f func(level string, format string, args ...any) bool) {
	if v, _ := ctx.Value(&log_ctx).(LogContext); v != nil {
		v.Range(f)
	}
	return
}

type SetLogCtx func(ctx context.Context, name string, limit int, levels []Level_t) context.Context

type log_ctx_middleware_t struct {
	Handler   http.Handler
	SetLogCtx SetLogCtx
	Levels    []Level_t
	Limit     int
}

func NewLogCtxMiddleware(next http.Handler, set SetLogCtx, limit int, levels []Level_t) http.Handler {
	self := &log_ctx_middleware_t{
		Handler:   next,
		SetLogCtx: set,
		Levels:    levels,
		Limit:     limit,
	}
	return self
}

func (self *log_ctx_middleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(self.SetLogCtx(r.Context(), uuid.New().String(), self.Limit, self.Levels))
	self.Handler.ServeHTTP(w, r)
}

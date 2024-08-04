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
	Get() []LogMsg_t
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

func (self *LogContext_t) Get() (res []LogMsg_t) {
	self.mx.Lock()
	res = self.data
	self.mx.Unlock()
	return
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

func GetLogContextPayload(ctx context.Context) (res []LogMsg_t) {
	if v, _ := ctx.Value(&log_ctx).(LogContext); v != nil {
		res = v.Get()
	}
	return
}

type SetCtx func(ctx context.Context, name string, limit int, levels []Level_t) context.Context

type errors_middleware_t struct {
	Handler http.Handler
	SetCtx  SetCtx
	Levels  []Level_t
	Limit   int
}

func NewErrorsMiddleware(next http.Handler, set SetCtx, limit int, levels []Level_t) http.Handler {
	self := &errors_middleware_t{
		Handler: next,
		SetCtx:  set,
		Levels:  levels,
		Limit:   limit,
	}
	return self
}

func (self *errors_middleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(self.SetCtx(r.Context(), uuid.New().String(), self.Limit, self.Levels))
	self.Handler.ServeHTTP(w, r)
}

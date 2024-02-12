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

type ErrorsContext interface {
	Name() string
	Set(level Level_t, format string, args ...any)
	Get() []string
	Reset()
}

type ErrorsContext_t struct {
	mx     sync.Mutex
	levels map[int64]Level_t
	name   string
	errors []string
}

func NewErrorsContext(name string, levels []Level_t) ErrorsContext {
	self := &ErrorsContext_t{
		name:   name,
		levels: map[int64]Level_t{},
	}
	for _, v := range levels {
		self.levels[v.Level] = v
	}
	return self
}

func (self *ErrorsContext_t) Name() string {
	return self.name
}

func (self *ErrorsContext_t) Set(level Level_t, format string, args ...any) {
	if _, ok := self.levels[level.Level]; ok {
		self.mx.Lock()
		self.errors = append(self.errors, format)
		self.mx.Unlock()
	}
}

func (self *ErrorsContext_t) Get() (res []string) {
	self.mx.Lock()
	res = self.errors
	self.mx.Unlock()
	return
}

func (self *ErrorsContext_t) Reset() {
	self.errors = nil
}

func SetErrorsContextNew(ctx context.Context, id string, levels []Level_t) context.Context {
	return context.WithValue(ctx, &log_ctx, NewErrorsContext(id, levels))
}

func SetErrorsContext(ctx context.Context, value ErrorsContext) context.Context {
	return context.WithValue(ctx, &log_ctx, value)
}

func GetErrorsContext(ctx context.Context) (value ErrorsContext) {
	value, _ = ctx.Value(&log_ctx).(ErrorsContext)
	return
}

func GetErrors(ctx context.Context) (res []string) {
	if v, _ := ctx.Value(&log_ctx).(ErrorsContext); v != nil {
		res = v.Get()
	}
	return
}

type SetCtx func(ctx context.Context, name string, levels []Level_t) context.Context

type errors_middleware_t struct {
	Handler http.Handler
	SetCtx  SetCtx
	Levels  []Level_t
}

func NewErrorsMiddleware(next http.Handler, set SetCtx, levels []Level_t) http.Handler {
	self := &errors_middleware_t{
		Handler: next,
		SetCtx:  set,
		Levels:  levels,
	}
	return self
}

func (self *errors_middleware_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(self.SetCtx(r.Context(), uuid.New().String(), self.Levels))
	self.Handler.ServeHTTP(w, r)
}

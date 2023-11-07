//
//
//

package log

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type context_key_t string

var context_key context_key_t = "log_ctx"

type ErrorsContext interface {
	Name() string
	Set(level Level_t, format string, args ...any)
	Get(out io.Writer)
	Reset()
}

type ErrorsContext_t struct {
	mx     sync.Mutex
	levels map[int64]Level_t
	name   string
	errors string
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
		if ix := strings.Index(format, " "); ix > 0 {
			self.errors = format[:ix]
		} else {
			self.errors = format
		}
		self.mx.Unlock()
	}
}

func (self *ErrorsContext_t) Get(out io.Writer) {
	self.mx.Lock()
	io.WriteString(out, self.errors)
	self.mx.Unlock()
	return
}

func (self *ErrorsContext_t) Reset() {
	self.errors = ""
}

func SetErrorsContextNew(ctx context.Context, id string, levels []Level_t) context.Context {
	return context.WithValue(ctx, context_key, NewErrorsContext(id, levels))
}

func SetErrorsContext(ctx context.Context, value ErrorsContext) context.Context {
	return context.WithValue(ctx, context_key, value)
}

func GetErrorsContext(ctx context.Context) (value ErrorsContext) {
	value, _ = ctx.Value(context_key).(ErrorsContext)
	return
}

func GetErrors(ctx context.Context, out io.Writer) {
	if v, _ := ctx.Value(context_key).(ErrorsContext); v != nil {
		v.Get(out)
	}
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

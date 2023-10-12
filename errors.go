//
//
//

package log

import (
	"context"
	"io"
	"strings"
	"sync"
)

type errors_context_t string

var errors_context errors_context_t = "log_ctx"

type ErrorsContext interface {
	Name() string
	Set(level string, format string, args ...any)
	Get(out io.Writer)
	Reset()
}

type ErrorsContext_t struct {
	mx     sync.Mutex
	name   string
	levels string
	errors string
}

func NewErrorsContext(name string, levels string) ErrorsContext {
	return &ErrorsContext_t{
		name:   name,
		levels: levels,
	}
}

func (self *ErrorsContext_t) Name() string {
	return self.name
}

func (self *ErrorsContext_t) Set(level string, format string, args ...any) {
	if strings.Contains(self.levels, level) {
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

func SetErrorsContextNew(ctx context.Context, id string, levels string) context.Context {
	return context.WithValue(ctx, errors_context, NewErrorsContext(id, levels))
}

func SetErrorsContext(ctx context.Context, value ErrorsContext) context.Context {
	return context.WithValue(ctx, errors_context, value)
}

func GetErrorsContext(ctx context.Context) (value ErrorsContext) {
	value, _ = ctx.Value(errors_context).(ErrorsContext)
	return
}

func GetErrors(ctx context.Context, out io.Writer) {
	if v, _ := ctx.Value(errors_context).(ErrorsContext); v != nil {
		v.Get(out)
	}
}

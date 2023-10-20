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

func (self *ErrorsContext_t) Set(level Level_t, format string, args ...any) {
	if strings.Contains(self.levels, level.Name) {
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

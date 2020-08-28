//
//
//

package log

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

var DT = &DT_t{Format: "2006-01-02 15:04:05"}

var std = NewLogger("stderr", LOG_TRACE, NewStderr(DT))

type DT_t struct {
	Format string
}

func (self *DT_t) Prefix() string {
	return time.Now().Format(self.Format)
}

type DTFL_t struct {
	Format string
	Depth  int
}

func (self *DTFL_t) Prefix() string {
	_, file, line, ok := runtime.Caller(self.Depth)
	if ok {
		return fmt.Sprintf("%s %s:%d", time.Now().Format(self.Format), path.Base(file), line)
	}
	return time.Now().Format(self.Format)
}

type NoWriter_t struct{}

func (NoWriter_t) WriteLevel(level string, format string, args ...interface{}) (int, error) {
	return 0, nil
}

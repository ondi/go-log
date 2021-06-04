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

var std = NewLogger("stderr", LOG_TRACE, NewStderr(&DTFL_t{Format: "2006-01-02 15:04:05", Depth: 4}))

var NoWriter = NoWriter_t{}

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

func (NoWriter_t) Close() error {
	return nil
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

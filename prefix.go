//
//
//

package log

import (
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type Prefixer interface {
	Prefix() string
}

type DT_t struct {
	Format string
}

func (self *DT_t) Prefix() string {
	var b [64]byte
	return string(time.Now().AppendFormat(b[:0], self.Format))
}

type FL_t struct{}

func (self *FL_t) Prefix() string {
	var ok bool
	var prev_path, next_path string
	var prev_line, next_line int
	_, prev_path, prev_line, ok = runtime.Caller(1)
	for i := 2; i < 100; i++ {
		if _, next_path, next_line, ok = runtime.Caller(i); ok {
			if filepath.Dir(prev_path) != filepath.Dir(next_path) {
				prev_path, prev_line = next_path, next_line
				break
			}
		} else {
			break
		}
	}
	return filepath.Base(prev_path) + ":" + strconv.FormatInt(int64(prev_line), 10)
}

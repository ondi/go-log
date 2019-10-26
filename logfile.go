//
// Log rotator
//

package log

import "os"
import "fmt"
import "sync"
import "time"

type RotateLogWriter struct {
	mx sync.Mutex
	fp * os.File
	datetime func() string
	filename string
	max_bytes int
	curr_bytes int
	backup_count int
}

func (self * RotateLogWriter) Write(level string, format string, args ...interface{}) {
	self.mx.Lock()
	defer self.mx.Unlock()
	n, err := fmt.Fprintf(self.fp, self.datetime() + level + " " + format + "\n", args...)
	if err != nil {
		return
	}
	self.curr_bytes += n
	if self.curr_bytes >= self.max_bytes {
		self.LogCycle()
	}
	return
}

func (self * RotateLogWriter) LogCycle() (err error) {
	if self.fp != nil {
		self.fp.Close()
	}
	os.Remove(fmt.Sprintf("%s.%d", self.filename, self.backup_count))
	for i := self.backup_count; i > 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", self.filename, i - 1), fmt.Sprintf("%s.%d", self.filename, i))
	}
	if self.backup_count > 0 {
		os.Rename(self.filename, fmt.Sprintf("%s.%d", self.filename, 1))
	} else {
		os.Remove(self.filename)
	}
	self.curr_bytes = 0
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY | os.O_CREATE /*| os.O_APPEND*/, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return
}

func NewRotateLogWriter(filename string, datetime string, max_bytes int, backup_count int) (self * RotateLogWriter) {
	self = &RotateLogWriter{filename: filename, max_bytes: max_bytes, backup_count: backup_count}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		self.datetime = func() string {return ""}
	}
	self.LogCycle()
	return
}

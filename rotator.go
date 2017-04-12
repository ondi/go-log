//
// Log rotator
//

package log

import "os"
import "fmt"
import "sync"

type RotateLogWriter struct {
	mx sync.Mutex
	fp * os.File
	filename string
	max_bytes int
	curr_bytes int
	backup_count int
}

func (self * RotateLogWriter) Write(m []byte) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.curr_bytes + len(m) >= self.max_bytes {
		self.LogCycle()
	}
	n, err = self.fp.Write(m)
	self.curr_bytes += n
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

func NewRotateLogWriter(filename string, max_bytes int, backup_count int) (self * RotateLogWriter) {
	self = &RotateLogWriter{filename: filename, max_bytes: max_bytes, backup_count: backup_count}
	self.LogCycle()
	return
}

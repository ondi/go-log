//
// Log rotator
//

package log

import "os"
import "fmt"
import "sync"
import "time"

type File_t struct {
	mx sync.Mutex
	fp * os.File
	datetime func() string
	filename string
	max_bytes int
	curr_bytes int
	backup_count int
}

func NewFile(filename string, datetime string, max_bytes int, backup_count int) (self * File_t, err error) {
	self = &File_t{filename: filename, max_bytes: max_bytes, backup_count: backup_count}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string {return time.Now().Format(datetime)}
	} else {
		self.datetime = func() string {return ""}
	}
	err = self.Cycle()
	return
}

func (self * File_t) Write(level string, format string, args ...interface{}) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if n, err = fmt.Fprintf(self.fp, self.datetime() + level + " " + format + "\n", args...); err != nil {
		return
	}
	self.curr_bytes += n
	if self.curr_bytes >= self.max_bytes {
		self.Cycle()
	}
	return
}

func (self * File_t) Cycle() (err error) {
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

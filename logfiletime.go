//
// Log rotator
//

package log

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type FileTime_t struct {
	mx           sync.Mutex
	fp           *os.File
	datetime     func() string
	filename     string
	truncate     time.Duration
	last_date    time.Time
	backup_count int
	files        []string
}

func NewFileTime(filename string, datetime string, truncate time.Duration, backup_count int) (self *FileTime_t, err error) {
	self = &FileTime_t{filename: filename, truncate: truncate, backup_count: backup_count}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string { return time.Now().Format(datetime) }
	} else {
		self.datetime = func() string { return "" }
	}
	err = self.__cycle()
	return
}

func (self *FileTime_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(self, self.datetime()+level+" "+format+"\n", args...)
}

func (self *FileTime_t) Write(p []byte) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	ts := time.Now()
	if !self.last_date.Equal(ts.Truncate(self.truncate)) {
		self.__cycle()
	}
	if n, err = self.fp.Write(p); err != nil {
		return
	}
	return
}

func (self *FileTime_t) __cycle() (err error) {
	if self.fp != nil {
		self.fp.Close()
		os.Rename(self.filename, fmt.Sprintf("%s.%s", self.filename, self.last_date.Format("2006-01-02T15:04:05")))
	}
	self.last_date = time.Now().Truncate(self.truncate)
	self.files = append(self.files, fmt.Sprintf("%s.%s", self.filename, self.last_date.Format("2006-01-02T15:04:05")))
	if len(self.files) > self.backup_count {
		os.Remove(self.files[0])
		self.files = self.files[1:]
	}
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY|os.O_CREATE /*|os.O_APPEND*/, 0644)
	return
}

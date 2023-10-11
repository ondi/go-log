//
// Log rotator
//

package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var FileTime = "20060102150405"

type FileTime_t struct {
	mx           sync.Mutex
	last_date    time.Time
	prefix       []Prefixer
	out          *os.File
	filename     string
	files        []string
	truncate     time.Duration
	backup_count int
	cycle        int
}

func NewFileTime(filename string, prefix []Prefixer, truncate time.Duration, backup_count int) (self *FileTime_t, err error) {
	self = &FileTime_t{
		prefix:       prefix,
		filename:     filename,
		truncate:     truncate,
		backup_count: backup_count,
		last_date:    time.Now(),
	}
	err = self.__cycle(self.last_date)
	return
}

func (self *FileTime_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	for _, v := range self.prefix {
		v.Prefix(self.out)
	}
	io.WriteString(self.out, level)
	io.WriteString(self.out, " ")
	n, err = fmt.Fprintf(self, format, args...)
	io.WriteString(self.out, "\n")
	return
}

func (self *FileTime_t) Write(p []byte) (int, error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	ts := time.Now()
	if !self.last_date.Equal(ts.Truncate(self.truncate)) {
		self.__cycle(ts)
		self.last_date = ts.Truncate(self.truncate)
	}
	return self.out.Write(p)
}

func (self *FileTime_t) __cycle(ts time.Time) (err error) {
	if self.out != nil {
		self.cycle++
		self.out.Close()
		backlog_file := fmt.Sprintf("%s.%d.%s", self.filename, self.cycle, ts.Format(FileTime))
		os.Rename(self.filename, backlog_file)
		self.files = append(self.files, backlog_file)
	}
	if len(self.files) > self.backup_count {
		os.Remove(self.files[0])
		self.files = self.files[1:]
	}
	self.out, err = os.OpenFile(self.filename, os.O_WRONLY|os.O_CREATE /*|os.O_APPEND*/, 0644)
	return
}

func (self *FileTime_t) Close() (err error) {
	if self.out != nil {
		err = self.out.Close()
	}
	return
}

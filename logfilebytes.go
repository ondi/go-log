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

var FileBytesFormat = "20060102150405"

type FileBytes_t struct {
	mx           sync.Mutex
	fp           *os.File
	datetime     DateTime
	filename     string
	bytes_limit  int
	bytes_count  int
	last_date    time.Time
	backup_count int
	files        []string
}

func NewFileBytes(filename string, datetime DateTime, bytes_limit int, backup_count int) (self *FileBytes_t, err error) {
	self = &FileBytes_t{
		datetime:     datetime,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
	}
	err = self.__cycle()
	return
}

func (self *FileBytes_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	dt := self.datetime.DateTime()
	if len(dt) > 0 {
		dt += " "
	}
	return fmt.Fprintf(self, dt+level+" "+format+"\n", args...)
}

func (self *FileBytes_t) Write(p []byte) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if n, err = self.fp.Write(p); err != nil {
		return
	}
	self.bytes_count += n
	if self.bytes_count >= self.bytes_limit {
		self.__cycle()
	}
	return
}

func (self *FileBytes_t) __cycle() (err error) {
	if self.fp != nil {
		self.fp.Close()
		os.Rename(self.filename, fmt.Sprintf("%s.%s", self.filename, self.last_date.Format(FileBytesFormat)))
	}
	self.bytes_count = 0
	self.last_date = time.Now()
	self.files = append(self.files, fmt.Sprintf("%s.%s", self.filename, self.last_date.Format(FileBytesFormat)))
	if len(self.files) > self.backup_count {
		os.Remove(self.files[0])
		self.files = self.files[1:]
	}
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY|os.O_CREATE /*|os.O_APPEND*/, 0644)
	return
}

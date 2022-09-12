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
	prefix       Prefixer
	filename     string
	bytes_limit  int
	bytes_count  int
	cycle        int
	backup_count int
	files        []string
}

func NewFileBytes(filename string, prefix Prefixer, bytes_limit int, backup_count int) (self *FileBytes_t, err error) {
	self = &FileBytes_t{
		prefix:       prefix,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
	}
	err = self.__cycle(time.Now())
	return
}

func (self *FileBytes_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	p := self.prefix.Prefix()
	if len(p) > 0 {
		p += " "
	}
	return fmt.Fprintf(self, p+level+" "+format+"\n", args...)
}

func (self *FileBytes_t) Write(p []byte) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if n, err = self.fp.Write(p); err != nil {
		return
	}
	self.bytes_count += n
	if self.bytes_count >= self.bytes_limit {
		self.__cycle(time.Now())
		self.bytes_count = 0
	}
	return
}

func (self *FileBytes_t) __cycle(ts time.Time) (err error) {
	if self.fp != nil {
		self.cycle++
		backlog_file := fmt.Sprintf("%s.%d.%s", self.filename, self.cycle, ts.Format(FileBytesFormat))
		self.fp.Close()
		os.Rename(self.filename, backlog_file)
		self.files = append(self.files, backlog_file)
	}
	if len(self.files) > self.backup_count {
		os.Remove(self.files[0])
		self.files = self.files[1:]
	}
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY|os.O_CREATE /*|os.O_APPEND*/, 0644)
	return
}

func (self *FileBytes_t) Close() error {
	return self.fp.Close()
}

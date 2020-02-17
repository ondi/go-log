//
// Log rotator
//

package log

import "os"
import "fmt"
import "sync"
import "time"

type FileBytes_t struct {
	mx           sync.Mutex
	fp           *os.File
	datetime     func() string
	filename     string
	bytes_limit  int
	bytes_count  int
	cycle_count  int
	backup_count int
	files        []string
}

func NewFileBytes(filename string, datetime string, bytes_limit int, backup_count int) (self *FileBytes_t, err error) {
	self = &FileBytes_t{filename: filename, bytes_limit: bytes_limit, backup_count: backup_count}
	if len(datetime) > 0 {
		datetime += " "
		self.datetime = func() string { return time.Now().Format(datetime) }
	} else {
		self.datetime = func() string { return "" }
	}
	err = self.__cycle()
	return
}

func (self *FileBytes_t) WriteLevel(level string, format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(self, self.datetime()+level+" "+format+"\n", args...)
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
		os.Rename(self.filename, fmt.Sprintf("%s.%d", self.filename, self.cycle_count))
	}
	self.cycle_count++
	self.bytes_count = 0
	self.files = append(self.files, fmt.Sprintf("%s.%d", self.filename, self.cycle_count))
	if len(self.files) > self.backup_count {
		os.Remove(self.files[0])
		self.files = self.files[1:]
	}
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY|os.O_CREATE /*|os.O_APPEND*/, 0644)
	return
}

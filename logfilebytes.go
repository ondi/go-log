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

var FileBytesFormat = "20060102150405"

type FileBytes_t struct {
	wg           sync.WaitGroup
	mx           sync.Mutex
	prefix       []Formatter
	out          *os.File
	filename     string
	files        []string
	bytes_limit  int
	bytes_count  int
	backup_count int
	cycle        int
	log_limit    int
	write_error  int
	write_total  int
}

func NewFileBytes(ts time.Time, filename string, prefix []Formatter, bytes_limit int, backup_count int, log_limit int) (Queue, error) {
	self := &FileBytes_t{
		prefix:       prefix,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
		log_limit:    log_limit,
	}
	return self, self.__cycle(ts)
}

func NewFileBytesQueue(queue_size int, writers int, ts time.Time, filename string, prefix []Formatter, bytes_limit int, backup_count int, log_limit int) (q Queue, err error) {
	self := &FileBytes_t{
		prefix:       prefix,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
		log_limit:    log_limit,
	}

	if err = self.__cycle(ts); err != nil {
		return
	}

	q = NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer(q)
	}

	return
}

func (self *FileBytes_t) writer(q Queue) (err error) {
	defer self.wg.Done()
	var msg [10]Msg_t
	for {
		n, oki := q.ReadLog(msg[:])
		if oki == -1 {
			return
		}
		for i := 0; i < n; i++ {
			if _, err = self.WriteLog(msg[i]); err != nil {
				q.WriteError(1)
			}
		}
	}
}

func (self *FileBytes_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.write_total++
	var w io.Writer
	if self.log_limit > 0 {
		w = &LimitWriter_t{Buf: self.out, Limit: self.log_limit}
	} else {
		w = self.out
	}
	for _, v := range self.prefix {
		n, err = v.FormatLog(w, m)
		self.bytes_count += n
	}
	n, err = io.WriteString(w, m.Level.Name)
	self.bytes_count += n
	n, err = io.WriteString(w, " ")
	self.bytes_count += n
	n, err = fmt.Fprintf(w, m.Format, m.Args...)
	self.bytes_count += n
	n, err = io.WriteString(self.out, "\n")
	self.bytes_count += n
	if self.bytes_count >= self.bytes_limit {
		self.__cycle(m.Level.Ts)
		self.bytes_count = 0
	}
	if err != nil {
		self.write_error++
	}
	return
}

func (self *FileBytes_t) ReadLog(p []Msg_t) (n int, oki int) {
	return 0, -1
}

func (self *FileBytes_t) WriteError(count int) {
}

func (self *FileBytes_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteError = self.write_error
	res.WriteTotal = self.write_total
	self.mx.Unlock()
	return
}

func (self *FileBytes_t) __cycle(ts time.Time) (err error) {
	if self.out != nil {
		self.cycle++
		backlog_file := fmt.Sprintf("%s.%d.%s", self.filename, self.cycle, ts.Format(FileBytesFormat))
		self.out.Close()
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

func (self *FileBytes_t) Close() (err error) {
	if self.out != nil {
		err = self.out.Close()
	}
	return
}

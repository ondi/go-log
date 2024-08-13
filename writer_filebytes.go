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

type WriterFileBytes_t struct {
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
	write_count  int
	write_error  int
	bulk_write   int
}

func NewWriterFileBytes(ts time.Time, filename string, prefix []Formatter, bytes_limit int, backup_count int, log_limit int) (Queue, error) {
	self := &WriterFileBytes_t{
		prefix:       prefix,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
		log_limit:    log_limit,
	}
	return self, self.__cycle(ts)
}

func NewWriterFileBytesQueue(queue_size int, writers int, ts time.Time, filename string, prefix []Formatter, bytes_limit int, backup_count int, log_limit int) (q Queue, err error) {
	self := &WriterFileBytes_t{
		prefix:       prefix,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
		log_limit:    log_limit,
		bulk_write:   16,
	}

	if err = self.__cycle(ts); err != nil {
		return
	}

	q = NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		q.WgAdd(1)
		go self.writer(q)
	}

	return
}

func (self *WriterFileBytes_t) writer(q Queue) (err error) {
	defer q.WgDone()
	msg := make([]Msg_t, self.bulk_write)
	for {
		n, ok := q.LogRead(msg)
		if !ok {
			return
		}
		for i := 0; i < n; i++ {
			if _, err = self.LogWrite(msg[i]); err != nil {
				q.WriteStat(n, n)
			} else {
				q.WriteStat(n, 0)
			}
		}
	}
}

func (self *WriterFileBytes_t) LogWrite(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.write_count++
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
	n, err = io.WriteString(w, m.Info.LevelName)
	self.bytes_count += n
	n, err = io.WriteString(w, " ")
	self.bytes_count += n
	n, err = fmt.Fprintf(w, m.Format, m.Args...)
	self.bytes_count += n
	n, err = io.WriteString(self.out, "\n")
	self.bytes_count += n
	if self.bytes_count >= self.bytes_limit {
		self.__cycle(m.Info.Ts)
		self.bytes_count = 0
	}
	if err != nil {
		self.write_error++
	}
	return
}

func (self *WriterFileBytes_t) LogRead(p []Msg_t) (n int, ok bool) {
	return
}

func (self *WriterFileBytes_t) WriteStat(count int, err int) {
}

func (self *WriterFileBytes_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteCount = self.write_count
	res.WriteError = self.write_error
	self.mx.Unlock()
	return
}

func (self *WriterFileBytes_t) WgAdd(int) {

}

func (self *WriterFileBytes_t) WgDone() {

}

func (self *WriterFileBytes_t) __cycle(ts time.Time) (err error) {
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

func (self *WriterFileBytes_t) Close() (err error) {
	self.mx.Lock()
	if self.out != nil {
		if err = self.out.Close(); err == nil {
			self.out = nil
		}
	}
	self.mx.Unlock()
	return
}

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
	mx              sync.Mutex
	prefix          []Formatter
	out             *os.File
	filename        string
	files           []string
	bytes_limit     int
	bytes_count     int
	backup_count    int
	cycle           int
	log_limit       int
	queue_write     int
	write_error_cnt int
	write_error_msg string
	bulk_write      int
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

func NewWriterFileBytesQueue(queue_size int, writers int, ts time.Time, filename string, prefix []Formatter, bytes_limit int, backup_count int, log_limit int) (Queue, error) {
	self := &WriterFileBytes_t{
		prefix:       prefix,
		filename:     filename,
		bytes_limit:  bytes_limit,
		backup_count: backup_count,
		log_limit:    log_limit,
		bulk_write:   16,
	}

	err := self.__cycle(ts)
	if err != nil {
		return nil, err
	}

	q := NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		q.WgAdd(1)
		go self.writer(q)
	}
	return q, err
}

func (self *WriterFileBytes_t) writer(q *Queue_t) (err error) {
	defer q.WgDone()
	for {
		msg, ok := q.LogRead(self.bulk_write)
		if !ok {
			return
		}
		for i := 0; i < len(msg); i++ {
			if _, err = self.LogWrite(msg[i]); err != nil {
				q.WriteError(1, err.Error())
			}
		}
	}
}

func (self *WriterFileBytes_t) LogWrite(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.queue_write++
	var w io.Writer
	if self.log_limit > 0 {
		w = &LimitWriter_t{Buf: self.out, Limit: self.log_limit}
	} else {
		w = self.out
	}
	for _, v := range self.prefix {
		n, err = v.FormatMessage(w, m)
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
		self.write_error_cnt++
		self.write_error_msg = err.Error()
	}
	return
}

func (self *WriterFileBytes_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.QueueWrite = self.queue_write
	res.WriteErrorCnt = self.write_error_cnt
	res.WriteErrorMsg = self.write_error_msg
	self.mx.Unlock()
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

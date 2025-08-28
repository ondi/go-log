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

type WriterFileTime_t struct {
	mx              sync.Mutex
	last_date       time.Time
	prefix          []Formatter
	out             *os.File
	filename        string
	files           []string
	truncate        time.Duration
	backup_count    int
	cycle           int
	log_limit       int
	queue_write     int
	write_error_cnt int
	write_error_msg string
	bulk_write      int
}

func NewWriterFileTime(ts time.Time, filename string, prefix []Formatter, truncate time.Duration, backup_count int, log_limit int) (Queue, error) {
	self := &WriterFileTime_t{
		prefix:       prefix,
		filename:     filename,
		truncate:     truncate,
		backup_count: backup_count,
		last_date:    ts,
		log_limit:    log_limit,
	}
	return self, self.__cycle(self.last_date)
}

func (self *WriterFileTime_t) LogWrite(msg []Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, m := range msg {
		self.queue_write++
		var w io.Writer
		if self.log_limit > 0 {
			w = &LimitWriter_t{Buf: self.out, Limit: self.log_limit}
		} else {
			w = self.out
		}
		if tr := m.Info.Ts.Truncate(self.truncate); !self.last_date.Equal(tr) {
			self.__cycle(m.Info.Ts)
			self.last_date = tr
		}
		for _, v := range self.prefix {
			v.FormatMessage(w, m)
		}
		n, err = fmt.Fprintf(w, m.Format, m.Args...)
		io.WriteString(self.out, "\n")
		if err != nil {
			self.write_error_cnt++
			self.write_error_msg = err.Error()
		}
	}
	return
}

func (self *WriterFileTime_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.QueueWrite = self.queue_write
	res.WriteErrorCnt = self.write_error_cnt
	res.WriteErrorMsg = self.write_error_msg
	self.mx.Unlock()
	return
}

func (self *WriterFileTime_t) Close() (err error) {
	self.mx.Lock()
	if self.out != nil {
		if err = self.out.Close(); err == nil {
			self.out = nil
		}
	}
	self.mx.Unlock()
	return
}

func (self *WriterFileTime_t) __cycle(ts time.Time) (err error) {
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
	self.out, err = os.OpenFile(self.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC /*|os.O_APPEND*/, 0644)
	return
}

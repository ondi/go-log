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
	wg           sync.WaitGroup
	mx           sync.Mutex
	last_date    time.Time
	prefix       []Formatter
	out          *os.File
	filename     string
	files        []string
	truncate     time.Duration
	backup_count int
	cycle        int
	log_limit    int
	write_error  int
	write_total  int
	bulk_write   int
}

func NewFileTime(ts time.Time, filename string, prefix []Formatter, truncate time.Duration, backup_count int, log_limit int) (Queue, error) {
	self := &FileTime_t{
		prefix:       prefix,
		filename:     filename,
		truncate:     truncate,
		backup_count: backup_count,
		last_date:    ts,
		log_limit:    log_limit,
	}
	return self, self.__cycle(self.last_date)
}

func NewFileTimeQueue(queue_size, writers int, ts time.Time, filename string, prefix []Formatter, truncate time.Duration, backup_count int, log_limit int) (q Queue, err error) {
	self := &FileTime_t{
		prefix:       prefix,
		filename:     filename,
		truncate:     truncate,
		backup_count: backup_count,
		last_date:    ts,
		log_limit:    log_limit,
		bulk_write:   16,
	}

	if err = self.__cycle(self.last_date); err != nil {
		return
	}

	q = NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer(q)
	}

	return
}

func (self *FileTime_t) writer(q Queue) (err error) {
	defer self.wg.Done()
	msg := make([]Msg_t, self.bulk_write)
	for {
		n, oki := q.ReadLog(msg)
		for i := 0; i < n; i++ {
			if _, err = self.WriteLog(msg[i]); err != nil {
				q.WriteError(1)
			}
		}
		if oki == -1 {
			return
		}
	}
}

func (self *FileTime_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.write_total++
	var w io.Writer
	if self.log_limit > 0 {
		w = &LimitWriter_t{Buf: self.out, Limit: self.log_limit}
	} else {
		w = self.out
	}
	if tr := m.Level.Ts.Truncate(self.truncate); !self.last_date.Equal(tr) {
		self.__cycle(m.Level.Ts)
		self.last_date = tr
	}
	for _, v := range self.prefix {
		v.FormatLog(w, m)
	}
	io.WriteString(w, m.Level.Name)
	io.WriteString(w, " ")
	n, err = fmt.Fprintf(w, m.Format, m.Args...)
	io.WriteString(self.out, "\n")
	if err != nil {
		self.write_error++
	}
	return
}

func (self *FileTime_t) ReadLog(p []Msg_t) (n int, oki int) {
	return 0, -1
}

func (self *FileTime_t) WriteError(count int) {
}

func (self *FileTime_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteError = self.write_error
	res.WriteTotal = self.write_total
	self.mx.Unlock()
	return
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

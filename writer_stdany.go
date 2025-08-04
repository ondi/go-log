//
//
//

package log

import (
	"fmt"
	"io"
	"sync"
)

type WriterStdany_t struct {
	mx              sync.Mutex
	prefix          []Formatter
	out             io.Writer
	log_limit       int
	queue_write     int
	write_error_cnt int
	write_error_msg string
	bulk_write      int
}

func NewWriterStdany(prefix []Formatter, out io.Writer, log_limit int) Queue {
	self := &WriterStdany_t{
		prefix:    prefix,
		out:       out,
		log_limit: log_limit,
	}
	return self
}

func NewWriterStdanyQueue(queue_size, writers int, prefix []Formatter, out io.Writer, log_limit int) Queue {
	self := &WriterStdany_t{
		prefix:     prefix,
		out:        out,
		log_limit:  log_limit,
		bulk_write: 16,
	}

	q := NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		q.WgAdd(1)
		go self.writer(q)
	}

	return q
}

func (self *WriterStdany_t) writer(q *Queue_t) (err error) {
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

func (self *WriterStdany_t) LogWrite(m Msg_t) (n int, err error) {
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
		v.FormatMessage(w, m)
	}
	n, err = fmt.Fprintf(w, m.Format, m.Args...)
	io.WriteString(self.out, "\n")
	if err != nil {
		self.write_error_cnt++
		self.write_error_msg = err.Error()
	}
	return
}

func (self *WriterStdany_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.QueueWrite = self.queue_write
	res.WriteErrorCnt = self.write_error_cnt
	res.WriteErrorMsg = self.write_error_msg
	self.mx.Unlock()
	return
}

func (self *WriterStdany_t) Close() error {
	return nil
}

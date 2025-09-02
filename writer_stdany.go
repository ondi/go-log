//
//
//

package log

import (
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

func (self *WriterStdany_t) LogWrite(msg []Msg_t) (n int, err error) {
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
		for _, v := range self.prefix {
			if n, err = v.FormatMessage(w, m); err != nil {
				self.write_error_cnt++
				self.write_error_msg = err.Error()
				return
			}
		}
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

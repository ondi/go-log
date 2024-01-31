//
//
//

package log

import (
	"fmt"
	"io"
	"sync"
)

type Stdany_t struct {
	mx          sync.Mutex
	prefix      []Formatter
	out         io.Writer
	log_limit   int
	write_error int
	write_total int
	bulk_write  int
}

func NewStdany(prefix []Formatter, out io.Writer, log_limit int) Queue {
	self := &Stdany_t{
		prefix:    prefix,
		out:       out,
		log_limit: log_limit,
	}
	return self
}

func NewStdanyQueue(queue_size, writers int, prefix []Formatter, out io.Writer, log_limit int) Queue {
	self := &Stdany_t{
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

func (self *Stdany_t) writer(q Queue) (err error) {
	defer q.WgDone()
	msg := make([]Msg_t, self.bulk_write)
	for {
		n, ok := q.ReadLog(msg)
		if !ok {
			return
		}
		for i := 0; i < n; i++ {
			if _, err = self.WriteLog(msg[i]); err != nil {
				q.WriteError(1)
			}
		}
	}
}

func (self *Stdany_t) WriteLog(m Msg_t) (n int, err error) {
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

func (self *Stdany_t) ReadLog(p []Msg_t) (n int, ok bool) {
	return
}

func (self *Stdany_t) WriteError(count int) {

}

func (self *Stdany_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteError = self.write_error
	res.WriteTotal = self.write_total
	self.mx.Unlock()
	return
}

func (self *Stdany_t) WgAdd(int) {

}

func (self *Stdany_t) WgDone() {

}

func (self *Stdany_t) Close() error {
	return nil
}

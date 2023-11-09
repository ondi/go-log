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
	wg          sync.WaitGroup
	mx          sync.Mutex
	prefix      []Formatter
	out         io.Writer
	write_error int
	write_total int
}

func NewStdany(prefix []Formatter, out io.Writer) Queue {
	self := &Stdany_t{
		prefix: prefix,
		out:    out,
	}
	return self
}

func NewStdanyQueue(queue_size, writers int, prefix []Formatter, out io.Writer) Queue {
	self := &Stdany_t{
		prefix: prefix,
		out:    out,
	}

	q := NewQueue(queue_size)
	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer(q)
	}

	return q
}

func (self *Stdany_t) writer(q Queue) (err error) {
	defer self.wg.Done()

	for {
		ms, oki := q.ReadLog(1)
		if oki == -1 {
			return
		}
		for _, v := range ms {
			if _, err = self.WriteLog(v); err != nil {
				self.write_error++
			}
		}
	}
}

func (self *Stdany_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.write_total++
	for _, v := range self.prefix {
		v.FormatLog(self.out, m)
	}
	io.WriteString(self.out, m.Level.Name)
	io.WriteString(self.out, " ")
	n, err = fmt.Fprintf(self.out, m.Format, m.Args...)
	io.WriteString(self.out, "\n")
	return
}

func (self *Stdany_t) ReadLog(count int) (out []Msg_t, oki int) {
	return nil, -1
}

func (self *Stdany_t) WriteError(count int) {
}

func (self *Stdany_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteError, res.WriteTotal = self.write_error, self.write_total
	self.mx.Unlock()
	return
}

func (self *Stdany_t) Close() error {
	return nil
}

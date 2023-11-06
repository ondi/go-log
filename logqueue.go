//
//
//

package log

import (
	"fmt"
	"sync"

	"github.com/ondi/go-queue"
)

type queue_t struct {
	mx          sync.Mutex
	q           queue.Queue[Msg_t]
	queue_error int
	write_error int
}

func NewQueue(limit int) Queue {
	self := &queue_t{}
	self.q = queue.NewOpen[Msg_t](&self.mx, limit)
	return self
}

func (self *queue_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	if n = self.q.PushBackNoLock(m); n != 0 {
		self.queue_error++
		err = fmt.Errorf("OVERFLOW: %+v", self.__size())
	}
	self.mx.Unlock()
	return
}

func (self *queue_t) ReadLog(count int) (out []Msg_t, oki int) {
	var m Msg_t
	self.mx.Lock()
	for i := 0; i < count; i++ {
		if m, oki = self.q.PopFront(); oki == 0 {
			out = append(out, m)
		} else {
			break
		}
		if self.q.Size() == 0 {
			break
		}
	}
	self.mx.Unlock()
	return
}

func (self *queue_t) WriteError(count int) {
	self.mx.Lock()
	self.write_error += count
	self.mx.Unlock()
}

func (self *queue_t) __size() (res QueueSize_t) {
	res.Limit = self.q.Limit()
	res.Size = self.q.Size()
	res.Readers = self.q.Readers()
	res.Writers = self.q.Writers()
	res.QueueError = self.queue_error
	res.WriteError = self.write_error
	return
}

func (self *queue_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res = self.__size()
	self.mx.Unlock()
	return
}

func (self *queue_t) Close() (err error) {
	self.mx.Lock()
	self.q.Close()
	self.mx.Unlock()
	return
}

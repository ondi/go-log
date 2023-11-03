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
	mx sync.Mutex
	q  queue.Queue[Msg_t]
}

func NewQueue(limit int) Queue {
	self := &queue_t{}
	self.q = queue.NewOpen[Msg_t](&self.mx, limit)
	return self
}

func (self *queue_t) WriteLog(m Msg_t) (n int, err error) {
	self.mx.Lock()
	if n = self.q.PushBackNoLock(m); n != 0 {
		err = fmt.Errorf("OVERFLOW: limit=%v, size=%v, readers=%v, writers=%v", self.q.Limit(), self.q.Size(), self.q.Readers(), self.q.Writers())
	}
	self.mx.Unlock()
	return
}

func (self *queue_t) ReadLog(count int) (out []Msg_t, oki int) {
	self.mx.Lock()
	var m Msg_t
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

func (self *queue_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.Limit, res.Size, res.Readers, res.Writers = self.q.Limit(), self.q.Size(), self.q.Readers(), self.q.Writers()
	self.mx.Unlock()
	return
}

func (self *queue_t) Close() (err error) {
	self.mx.Lock()
	self.q.Close()
	self.mx.Unlock()
	return
}

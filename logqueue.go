//
//
//

package log

import (
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
	n = self.q.PushBackNoWait(m)
	self.mx.Unlock()
	return
}

func (self *queue_t) ReadLog(count int) (out []Msg_t, oki int) {
	self.mx.Lock()
	var m Msg_t
	for i := 0; i < count; i++ {
		m, oki = self.q.PopFront()
		if oki == 0 {
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

func (self *queue_t) Close() (err error) {
	self.mx.Lock()
	self.q.Close()
	self.mx.Unlock()
	return
}

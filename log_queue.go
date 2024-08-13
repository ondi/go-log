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
	wg          sync.WaitGroup
	mx          sync.Mutex
	q           queue.Queue[Msg_t]
	queue_write int
	queue_error int
	queue_read  int
	write_count int
	write_error int
}

func NewQueue(limit int) Queue {
	self := &queue_t{}
	self.q = queue.NewOpen[Msg_t](&self.mx, limit)
	return self
}

func (self *queue_t) LogWrite(m Msg_t) (n int, err error) {
	self.mx.Lock()
	self.queue_write++
	if self.q.PushBackNoLock(m) == false {
		self.queue_error++
		err = fmt.Errorf("QUEUE WRITE")
	}
	self.mx.Unlock()
	return
}

func (self *queue_t) LogRead(p []Msg_t) (n int, ok bool) {
	var m Msg_t
	self.mx.Lock()
	for n < len(p) {
		if m, ok = self.q.PopFront(); ok {
			p[n] = m
			n++
		} else {
			break
		}
		if self.q.Size() == 0 {
			break
		}
	}
	self.queue_read += n
	self.mx.Unlock()
	return
}

func (self *queue_t) WriteStat(count int, err int) {
	self.mx.Lock()
	self.write_count += count
	self.write_error += err
	self.mx.Unlock()
}

func (self *queue_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.Limit = self.q.Limit()
	res.Size = self.q.Size()
	res.Readers = self.q.Readers()
	res.Writers = self.q.Writers()
	res.QueueWrite = self.queue_write
	res.QueueError = self.queue_error
	res.QueueRead = self.queue_read
	res.WriteCount = self.write_count
	res.WriteError = self.write_error
	self.mx.Unlock()
	return
}

func (self *queue_t) WgAdd(n int) {
	self.wg.Add(n)
}

func (self *queue_t) WgDone() {
	self.wg.Done()
}

func (self *queue_t) Close() (err error) {
	self.mx.Lock()
	self.q.Close()
	self.mx.Unlock()
	self.wg.Wait()
	return
}

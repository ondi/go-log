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
	q           queue.Queue[LogMsg_t]
	queue_error int
	write_error int
	write_total int
	read_total  int
}

func NewQueue(limit int) Queue {
	self := &queue_t{}
	self.q = queue.NewOpen[LogMsg_t](&self.mx, limit)
	return self
}

func (self *queue_t) WriteLog(m LogMsg_t) (n int, err error) {
	self.mx.Lock()
	self.write_total++
	if self.q.PushBackNoLock(m) == false {
		self.queue_error++
		err = fmt.Errorf("QUEUE WRITE")
	}
	self.mx.Unlock()
	return
}

func (self *queue_t) ReadLog(p []LogMsg_t) (n int, ok bool) {
	var m LogMsg_t
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
	self.read_total += n
	self.mx.Unlock()
	return
}

func (self *queue_t) WriteError(count int) {
	self.mx.Lock()
	self.write_error += count
	self.mx.Unlock()
}

func (self *queue_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.Limit = self.q.Limit()
	res.Size = self.q.Size()
	res.Readers = self.q.Readers()
	res.Writers = self.q.Writers()
	res.QueueError = self.queue_error
	res.WriteError = self.write_error
	res.WriteTotal = self.write_total
	res.ReadTotal = self.read_total
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

//
//
//

package log

import (
	"errors"
	"sync"

	"github.com/ondi/go-queue"
)

var ERROR_OVERFLOW = errors.New("QUEUE OVERFLOW")

type queue_t struct {
	wg              sync.WaitGroup
	mx              sync.Mutex
	q               queue.Queue[Msg_t]
	queue_write     int
	queue_read      int
	queue_overflow  int
	write_error_cnt int
	write_error_msg string
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
		self.queue_overflow++
		err = ERROR_OVERFLOW
	}
	self.mx.Unlock()
	return
}

// bad design: messages stay in buffer forever and not garbage-collected
// LogRead(p []Msg_t) (n int, ok bool)
func (self *queue_t) LogRead(limit int) (res []Msg_t, ok bool) {
	var m Msg_t
	self.mx.Lock()
	for len(res) < limit {
		if m, ok = self.q.PopFront(); ok {
			res = append(res, m)
		} else {
			break
		}
		if self.q.Size() == 0 {
			break
		}
	}
	self.queue_read += len(res)
	self.mx.Unlock()
	return
}

func (self *queue_t) WriteError(count int, msg string) {
	self.mx.Lock()
	self.write_error_cnt += count
	self.write_error_msg = msg
	self.mx.Unlock()
}

func (self *queue_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.Limit = self.q.Limit()
	res.Size = self.q.Size()
	res.Readers = self.q.Readers()
	res.Writers = self.q.Writers()
	res.QueueWrite = self.queue_write
	res.QueueOverflow = self.queue_overflow
	res.QueueRead = self.queue_read
	res.WriteErrorCnt = self.write_error_cnt
	res.WriteErrorMsg = self.write_error_msg
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

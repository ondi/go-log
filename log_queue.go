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

type Queue_t struct {
	wg              sync.WaitGroup
	mx              sync.Mutex
	q               queue.Queue[Msg_t]
	queue_write     int
	queue_read      int
	queue_overflow  int
	write_error_cnt int
	write_error_msg string
}

func NewQueue(limit int, writers int, bulk_write int, w Queue) (self *Queue_t) {
	self = &Queue_t{}
	self.q = queue.NewOpen[Msg_t](&self.mx, limit)
	for i := 0; i < writers; i++ {
		self.wg.Add(1)
		go self.writer(bulk_write, w)
	}
	return self
}

func (self *Queue_t) writer(bulk_write int, w Queue) (err error) {
	defer self.wg.Done()
	for {
		msg, ok := self.LogRead(bulk_write)
		if !ok {
			return
		}
		if _, err = w.LogWrite(msg); err != nil {
			self.WriteError(1, err.Error())
		}
	}
}

func (self *Queue_t) LogWrite(msg []Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.queue_write += len(msg)
	for _, m := range msg {
		if self.q.PushBackNoLock(m) == false {
			self.queue_overflow++
			err = ERROR_OVERFLOW
			return
		}
	}
	return
}

// LogRead(p []Msg_t) (n int, ok bool) - bad design
// messages stay in buffer forever and not garbage-collected
func (self *Queue_t) LogRead(limit int) (res []Msg_t, ok bool) {
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

func (self *Queue_t) WriteError(count int, msg string) {
	self.mx.Lock()
	self.write_error_cnt += count
	self.write_error_msg = msg
	self.mx.Unlock()
}

func (self *Queue_t) Size() (res QueueSize_t) {
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

func (self *Queue_t) Close() (err error) {
	self.mx.Lock()
	self.q.Close()
	self.mx.Unlock()
	self.wg.Wait()
	return
}

//
//
//

package log

import (
	"sync/atomic"
)

type WriterCounter_t struct {
	queue_write atomic.Int64
}

func NewWriterCounter() Queue {
	return &WriterCounter_t{}
}

func (self *WriterCounter_t) LogWrite(Msg_t) (n int, err error) {
	self.queue_write.Add(1)
	return
}

func (self *WriterCounter_t) Size() (res QueueSize_t) {
	res.QueueWrite = int(self.queue_write.Load())
	return
}

func (self *WriterCounter_t) Close() error {
	return nil
}

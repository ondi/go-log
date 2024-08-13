//
//
//

package log

import (
	"sync/atomic"
)

type WriterCounter_t struct {
	write_count atomic.Int64
}

func NewWriterCounter() Queue {
	return &WriterCounter_t{}
}

func (self *WriterCounter_t) LogWrite(Msg_t) (n int, err error) {
	self.write_count.Add(1)
	return
}

func (self *WriterCounter_t) LogRead(p []Msg_t) (n int, ok bool) {
	return
}

func (self *WriterCounter_t) WriteStat(count int, err int) {

}

func (self *WriterCounter_t) Size() (res QueueSize_t) {
	res.WriteCount = int(self.write_count.Load())
	return
}

func (self *WriterCounter_t) WgAdd(int) {

}

func (self *WriterCounter_t) WgDone() {

}

func (self *WriterCounter_t) Close() error {
	return nil
}

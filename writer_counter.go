//
//
//

package log

import (
	"sync"
)

type WriterCounter_t struct {
	mx          sync.Mutex
	write_total int
}

func NewWriterCounter() Queue {
	return &WriterCounter_t{}
}

func (self *WriterCounter_t) WriteLog(Msg_t) (n int, err error) {
	self.mx.Lock()
	self.write_total++
	self.mx.Unlock()
	return
}

func (self *WriterCounter_t) ReadLog(p []Msg_t) (n int, ok bool) {
	return
}

func (self *WriterCounter_t) WriteError(count int) {

}

func (self *WriterCounter_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteTotal = self.write_total
	self.mx.Unlock()
	return
}

func (self *WriterCounter_t) WgAdd(int) {

}

func (self *WriterCounter_t) WgDone() {

}

func (self *WriterCounter_t) Close() error {
	return nil
}

//
//
//

package log

import (
	"sync"
)

type LogCounter_t struct {
	mx          sync.Mutex
	write_total int
}

func NewLogCounter() Queue {
	return &LogCounter_t{}
}

func (self *LogCounter_t) WriteLog(Msg_t) (n int, err error) {
	self.mx.Lock()
	self.write_total++
	self.mx.Unlock()
	return
}

func (self *LogCounter_t) ReadLog(p []Msg_t) (n int, ok bool) {
	return
}

func (self *LogCounter_t) WriteError(count int) {

}

func (self *LogCounter_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteTotal = self.write_total
	self.mx.Unlock()
	return
}

func (self *LogCounter_t) WgAdd(int) {

}

func (self *LogCounter_t) WgDone() {

}

func (self *LogCounter_t) Close() error {
	return nil
}

func (self *LogCounter_t) Closed() bool {
	return true
}

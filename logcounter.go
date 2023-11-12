//
//
//

package log

import (
	"sync"
)

type LogCounter_t struct {
	mx    sync.Mutex
	total int
}

func NewLogCounter() Queue {
	return &LogCounter_t{}
}

func (self *LogCounter_t) WriteLog(Msg_t) (n int, err error) {
	self.mx.Lock()
	self.total++
	self.mx.Unlock()
	return
}

func (self *LogCounter_t) ReadLog(p []Msg_t) (n int, oki int) {
	return 0, -1
}

func (self *LogCounter_t) WriteError(count int) {
}

func (self *LogCounter_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteTotal = self.total
	self.mx.Unlock()
	return
}

func (self *LogCounter_t) Close() error {
	return nil
}

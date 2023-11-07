//
//
//

package log

import (
	"sync"
)

type LogCounter_t struct {
	mx     sync.Mutex
	errors int
}

func NewLogCounter() Queue {
	return &LogCounter_t{}
}

func (self *LogCounter_t) WriteLog(Msg_t) (n int, err error) {
	self.mx.Lock()
	self.errors++
	self.mx.Unlock()
	return
}

func (self *LogCounter_t) ReadLog(count int) (out []Msg_t, oki int) {
	return nil, -1
}

func (self *LogCounter_t) WriteError(count int) {
}

func (self *LogCounter_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.WriteError = self.errors
	self.mx.Unlock()
	return
}

func (self *LogCounter_t) Close() error {
	return nil
}

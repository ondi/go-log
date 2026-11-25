//
//
//

package log

import (
	"sync"
)

type WriterCounter_t struct {
	mx          sync.Mutex
	tags        map[string]map[string]int64
	queue_write int
}

func NewWriterCounter() Queue {
	return &WriterCounter_t{
		tags: map[string]map[string]int64{},
	}
}

func (self *WriterCounter_t) LogWrite(msg []Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range msg {
		found := 0
		level := LevelName(v.Info.Level)
		if self.tags[level] == nil {
			self.tags[level] = map[string]int64{}
		}
		for _, v2 := range v.Args {
			if temp, ok := v2.(Tag); ok {
				self.tags[level][temp.TagKey()]++
				found++
			}
		}
		if found == 0 {
			self.tags[level][""]++
		}
		self.queue_write++
	}
	return
}

func (self *WriterCounter_t) Range(f func(level string, key string, count int64) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for level, v1 := range self.tags {
		for key, v2 := range v1 {
			if f(level, key, v2) == false {
				return
			}
		}
	}
}

func (self *WriterCounter_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	defer self.mx.Unlock()
	res.QueueWrite = self.queue_write
	return
}

func (self *WriterCounter_t) Close() error {
	return nil
}

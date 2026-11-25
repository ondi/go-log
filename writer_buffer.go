//
//
//

package log

import (
	"sync"
	"time"

	"github.com/ondi/go-circular"
)

type WriterBuffer_t struct {
	mx          sync.Mutex
	data        *circular.List_t[Msg_t]
	queue_write int
	limit       int
}

func NewWriterBuffer(limit int) (self *WriterBuffer_t) {
	self = &WriterBuffer_t{
		data:  circular.New[Msg_t](limit),
		limit: limit,
	}
	return
}

func (self *WriterBuffer_t) LogWrite(m []Msg_t) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	for _, v := range m {
		if self.data.Size() >= self.limit {
			self.data.PopFront()
		}
		self.data.PushBack(v)
		self.queue_write++
	}
	return
}

func (self *WriterBuffer_t) Size() (res QueueSize_t) {
	self.mx.Lock()
	res.QueueWrite = self.queue_write
	self.mx.Unlock()
	return
}

func (self *WriterBuffer_t) Close() (err error) {
	return
}

func (self *WriterBuffer_t) BufferRange(f func(ts time.Time, file string, line int, level_id int64, format string, args ...any) bool) {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.RangeFront(func(m Msg_t) bool {
		return f(m.Info.Ts, m.Info.File, m.Info.Line, m.Info.Level, m.Format, m.Args...)
	})
}

func (self *WriterBuffer_t) BufferReset() {
	self.mx.Lock()
	defer self.mx.Unlock()
	self.data.Reset()
}

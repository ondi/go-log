//
//
//

package log

import "sync/atomic"

type WriterContext_t struct {
	queue_write atomic.Int64
}

func NewWriterContext() (self *WriterContext_t) {
	self = &WriterContext_t{}
	return
}

func (self *WriterContext_t) LogWrite(m []Msg_t) (n int, err error) {
	for _, v := range m {
		if l := GetLogValue(v.Ctx); l != nil {
			if q := l.GetQueue(); q != nil {
				q.LogWrite(m)
				self.queue_write.Add(1)
			}
		}
	}
	return
}

func (self *WriterContext_t) Size() (res QueueSize_t) {
	res.QueueWrite = int(self.queue_write.Load())
	return
}

func (self *WriterContext_t) Close() (err error) {
	return
}

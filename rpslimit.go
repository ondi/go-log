//
//
//

package log

import (
	"sync"
	"time"

	ttl_cache "github.com/ondi/go-ttl-cache"
)

type Rps interface {
	Overflow(time.Time) bool
}

type NoRps_t struct{}

func (NoRps_t) Overflow(time.Time) bool {
	return false
}

type Rps_t struct {
	mx        sync.Mutex
	c         *ttl_cache.Cache_t
	evict     time.Duration
	count     int
	rps_limit int
}

/*
example 1000 rps:
ttl = 1s
evict = 0.05s
rps_limit=1000
*/
func NewRps(ttl time.Duration, evict time.Duration, rps_limit int) (self *Rps_t) {
	self = &Rps_t{}
	self.c = ttl_cache.New(1<<32-1, ttl, self.__evict)
	self.evict = evict
	self.rps_limit = rps_limit
	return
}

func (self *Rps_t) __evict(key interface{}, value interface{}) {
	self.count -= value.(int)
}

func (self *Rps_t) add(ts time.Time) (res int) {
	self.mx.Lock()
	self.c.Write(
		ts,
		ts.Truncate(self.evict),
		func() interface{} {
			return 1
		},
		func(prev interface{}) interface{} {
			return prev.(int) + 1
		},
	)
	self.count++
	res = self.count
	self.mx.Unlock()
	return
}

func (self *Rps_t) Overflow(ts time.Time) bool {
	if self.add(ts) > self.rps_limit {
		return true
	}
	return false
}

func (self *Rps_t) Size() (res int) {
	self.mx.Lock()
	res = self.c.Size()
	self.mx.Unlock()
	return
}

func (self *Rps_t) Count() (res int) {
	self.mx.Lock()
	res = self.count
	self.mx.Unlock()
	return
}

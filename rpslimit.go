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
	Add(time.Time) (ok bool)
}

type NoRps_t struct{}

func (NoRps_t) Add(time.Time) bool {
	return true
}

type Rps_t struct {
	mx        sync.Mutex
	c         *ttl_cache.Cache_t
	truncate  time.Duration
	count     int
	rps_limit int
}

/*
example 1000 rps:
ttl = time.Second
buckets = 100
rps_limit=1000
*/
func NewRps(ttl time.Duration, buckets int64, rps_limit int) (self *Rps_t) {
	self = &Rps_t{}
	self.c = ttl_cache.New(1<<32-1, ttl, self.__evict)
	self.truncate = ttl / time.Duration(buckets)
	self.rps_limit = rps_limit
	return
}

func (self *Rps_t) __evict(key interface{}, value interface{}) {
	self.count -= value.(int)
}

func (self *Rps_t) Add(ts time.Time) (ok bool) {
	self.mx.Lock()
	self.c.Create(
		ts,
		ts.Truncate(self.truncate),
		func() interface{} {
			if self.count == self.rps_limit {
				return 0
			}
			ok = true
			self.count++
			return 1
		},
		func(prev interface{}) interface{} {
			if self.count == self.rps_limit {
				return prev
			}
			ok = true
			self.count++
			return prev.(int) + 1
		},
	)
	self.mx.Unlock()
	return
}

func (self *Rps_t) Size(ts time.Time) (res int) {
	self.mx.Lock()
	res = self.c.Size(ts)
	self.mx.Unlock()
	return
}

func (self *Rps_t) Count() (res int) {
	self.mx.Lock()
	res = self.count
	self.mx.Unlock()
	return
}

//
//
//

package log

import (
	"sync"
	"time"

	"github.com/ondi/go-cache"
)

type Rps interface {
	Add(ts time.Time) (ok bool)
}

type NoRps_t struct{}

func (NoRps_t) Add(time.Time) bool {
	return true
}

type Rps_t struct {
	mx        sync.Mutex
	c         *cache.Cache_t[time.Time, int]
	ttl       time.Duration
	truncate  time.Duration
	count     int
	buckets   int
	rps_limit int
}

/*
example 1000 rps:
ttl = time.Second
buckets = 100
rps_limit=1000
*/
func NewRps(ttl time.Duration, buckets int64, rps_limit int) (self *Rps_t) {
	self = &Rps_t{
		c: cache.New[time.Time, int](),
		ttl: ttl,
		truncate: ttl / time.Duration(buckets),
		buckets: int(buckets),
		rps_limit: rps_limit,
	}
	return
}

func (self *Rps_t) __flush(ts time.Time) {
	for it := self.c.Front(); it != self.c.End(); it = it.Next() {
		if ts.After(it.Key) || self.c.Size() > self.buckets {
			self.c.Remove(it.Key)
			self.count -= it.Value
		} else {
			return
		}
	}
}

func (self *Rps_t) Add(ts time.Time) bool {
	self.mx.Lock()
	self.__flush(ts)
	if self.count == self.rps_limit {
		self.mx.Unlock()
		return false
	}
	it, _ := self.c.CreateBack(
		ts.Add(self.ttl).Truncate(self.truncate),
		func() int { return 0 },
	)
	it.Value++
	self.count++
	self.mx.Unlock()
	return true
}

func (self *Rps_t) Size(ts time.Time) (s1 int) {
	self.mx.Lock()
	self.__flush(ts)
	s1 = self.c.Size()
	self.mx.Unlock()
	return
}

func (self *Rps_t) Count(id string) (res int) {
	self.mx.Lock()
	res = self.count
	self.mx.Unlock()
	return
}

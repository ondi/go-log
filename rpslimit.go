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
	Add(id string, ts time.Time) (ok bool)
}

type NoRps_t struct{}

func (NoRps_t) Add(string, time.Time) bool {
	return true
}

type RpsKey_t struct {
	Id string
	Ts time.Time
}

type Rps_t struct {
	mx        sync.Mutex
	c         *cache.Cache_t[RpsKey_t, int]
	count     map[string]int
	ttl       time.Duration
	truncate  time.Duration
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
		c: cache.New[RpsKey_t, int](),
		count: map[string]int{},
		ttl: ttl,
		truncate: ttl / time.Duration(buckets),
		rps_limit: rps_limit,
	}
	return
}

func (self *Rps_t) Flush(ts time.Time) {
	for it := self.c.Front(); it != self.c.End(); it = it.Next() {
		if ts.Before(it.Key.Ts) {
			return
		}
		self.c.Remove(it.Key)
		if self.count[it.Key.Id] == it.Value {
			delete(self.count, it.Key.Id)
		} else {
			self.count[it.Key.Id] -= it.Value
		}
	}
}

func (self *Rps_t) Add(id string, ts time.Time) bool {
	self.mx.Lock()
	self.Flush(ts)
	if self.count[id] == self.rps_limit {
		self.mx.Unlock()
		return false
	}
	it, _ := self.c.CreateBack(
		RpsKey_t{
			Id: id,
			Ts: ts.Add(self.ttl).Truncate(self.truncate),
		},
		func() int { return 0 },
	)
	it.Value = it.Value + 1
	self.count[id]++
	self.mx.Unlock()
	return true
}

func (self *Rps_t) Size(ts time.Time) (s1 int, s2 int) {
	self.mx.Lock()
	self.Flush(ts)
	s1 = self.c.Size()
	s2 = len(self.count)
	self.mx.Unlock()
	return
}

func (self *Rps_t) Count(id string) (res int) {
	self.mx.Lock()
	res = self.count[id]
	self.mx.Unlock()
	return
}

//@File     cache_stat.go
//@Time     2025/5/28
//@Author   #Suyghur,

package internal

import (
	"sync/atomic"
	"time"

	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/timex"
)

const statInterval = time.Minute

// A CacheStat is used to stat the cache.
type (
	CacheStat struct {
		name string
		// export the fields to let the unit tests working,
		// reside in internal package, doesn't matter.
		Total   uint64
		Hit     uint64
		Miss    uint64
		DbFails uint64
	}
)

// NewCacheStat returns a CacheStat.
func NewCacheStat(name string) *CacheStat {
	s := &CacheStat{
		name: name,
	}

	gopool.Go(func() {
		ticker := timex.NewTicker(statInterval)
		defer ticker.Stop()

		s.statLoop(ticker)
	})

	return s
}

// IncrementTotal increments the total count.
func (s *CacheStat) IncrementTotal() {
	atomic.AddUint64(&s.Total, 1)
}

// IncrementHit increments the hit count.
func (s *CacheStat) IncrementHit() {
	atomic.AddUint64(&s.Hit, 1)
}

// IncrementMiss increments the miss count.
func (s *CacheStat) IncrementMiss() {
	atomic.AddUint64(&s.Miss, 1)
}

// IncrementDbFails increments the db fail count.
func (s *CacheStat) IncrementDbFails() {
	atomic.AddUint64(&s.DbFails, 1)
}

func (s *CacheStat) statLoop(ticker timex.Ticker) {
	for range ticker.Chan() {
		total := atomic.SwapUint64(&s.Total, 0)
		if total == 0 {
			continue
		}

		hit := atomic.SwapUint64(&s.Hit, 0)
		percent := 100 * float32(hit) / float32(total)
		miss := atomic.SwapUint64(&s.Miss, 0)
		dbf := atomic.SwapUint64(&s.DbFails, 0)
		logx.Statf("dbcache(%s) - qpm: %d, hit_ratio: %.1f%%, hit: %d, miss: %d, db_fails: %d", s.name, total, percent, hit, miss, dbf)
	}
}

//@File     period_limit_test.go
//@Time     2025/8/7
//@Author   #Suyghur,

package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeriodLimit_Take(t *testing.T) {
	testPeriodLimit(t)
}

func TestPeriodLimit_TakeWithAlign(t *testing.T) {
	testPeriodLimit(t, Align())
}

func TestPeriodLimit_RedisUnavailable(t *testing.T) {
	store := CreateRedis(t)
	const (
		seconds = 1
		quota   = 5
	)
	l := NewPeriodLimit(seconds, quota, store, "periodlimit")
	store.Close()
	val, err := l.Take("first")
	assert.NotNil(t, err)
	assert.Equal(t, 0, val)
}

func testPeriodLimit(t *testing.T, opts ...PeriodOption) {
	store := CreateRedis(t)

	const (
		seconds = 1
		total   = 100
		quota   = 5
	)
	l := NewPeriodLimit(seconds, quota, store, "periodlimit", opts...)
	var allowed, hitQuota, overQuota int
	for i := 0; i < total; i++ {
		val, err := l.Take("first")
		if err != nil {
			t.Error(err)
		}
		switch val {
		case Allowed:
			allowed++
		case HitQuota:
			hitQuota++
		case OverQuota:
			overQuota++
		default:
			t.Error("unknown status")
		}
	}

	assert.Equal(t, quota-1, allowed)
	assert.Equal(t, 1, hitQuota)
	assert.Equal(t, total-quota, overQuota)
}

func TestQuotaFull(t *testing.T) {
	store := CreateRedis(t)

	l := NewPeriodLimit(1, 1, store, "periodlimit")
	val, err := l.Take("first")
	assert.Nil(t, err)
	assert.Equal(t, HitQuota, val)
}

//@File     redis_test.go
//@Time     2025/5/29
//@Author   #Suyghur,

package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

// CreateRedis returns an in process Redis.
func CreateRedis(t *testing.T) *Redis {
	r, _ := CreateRedisWithCloseHandler(t)
	return r
}

// CreateRedisWithCloseHandler returns an in process Redis and a close function.
func CreateRedisWithCloseHandler(t *testing.T) (*Redis, func()) {
	rds := miniredis.RunT(t)
	return NewRedis(RedisConf{
		Host: rds.Addr(),
	}), rds.Close
}

func RunOnRedis(t *testing.T, fn func(rds *Redis)) {
	logx.Disable()
	s := miniredis.RunT(t)
	fn(NewRedis(RedisConf{
		Host: s.Addr(),
	}))
}

//@File     redis.go
//@Time     2023/10/31
//@Author   #Suyghur,

package dbtest

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"testing"
)

// CreateRedis returns an in process redis.Redis.
func CreateRedis(t *testing.T) *redis.Redis {
	r, _ := CreateRedisWithCloseHandler(t)
	return r
}

// CreateRedisWithCloseHandler returns an in process redis.Redis and a close function.
func CreateRedisWithCloseHandler(t *testing.T) (r *redis.Redis, handler func()) {
	rds := miniredis.RunT(t)
	return redis.MustNewRedis(redis.RedisConf{
		Host: rds.Addr(),
		Type: redis.NodeType,
	}), rds.Close
}

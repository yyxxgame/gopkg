//@File     lock_test.go
//@Time     2025/5/28
//@Author   #Suyghur,

package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stringx"
)

func TestRedisLock(t *testing.T) {
	testFn := func(ctx context.Context) func(rds *Redis) {
		return func(rds *Redis) {
			key := stringx.Rand()
			firstLock := NewLock(rds, key)
			firstLock.SetExpire(5)
			firstAcquire, err := firstLock.Acquire()
			assert.Nil(t, err)
			assert.True(t, firstAcquire)

			secondLock := NewLock(rds, key)
			secondLock.SetExpire(5)
			againAcquire, err := secondLock.Acquire()
			assert.Nil(t, err)
			assert.False(t, againAcquire)

			release, err := firstLock.Release()
			assert.Nil(t, err)
			assert.True(t, release)

			endAcquire, err := secondLock.Acquire()
			assert.Nil(t, err)
			assert.True(t, endAcquire)
		}
	}

	t.Run("normal", func(t *testing.T) {
		RunOnRedis(t, testFn(nil))
	})

	t.Run("withContext", func(t *testing.T) {
		RunOnRedis(t, testFn(context.Background()))
	})
}

func TestRedisLock_Expired(t *testing.T) {
	RunOnRedis(t, func(rds *Redis) {
		key := stringx.Rand()
		lock := NewLock(rds, key)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := lock.AcquireCtx(ctx)
		assert.NotNil(t, err)
	})

	RunOnRedis(t, func(rds *Redis) {
		key := stringx.Rand()
		lock := NewLock(rds, key)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := lock.ReleaseCtx(ctx)
		assert.NotNil(t, err)
	})
}

//@File     lock.go
//@Time     2025/5/28
//@Author   #Suyghur,

package redis

import (
	"context"
	_ "embed"
	"errors"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	v9rds "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stringx"
)

const (
	randomLen       = 16
	tolerance       = 500 // milliseconds
	millisPerSecond = 1000
)

var (
	//go:embed internal/lockscript.lua
	lockLuaScript string
	lockScript    = v9rds.NewScript(lockLuaScript)

	//go:embed internal/delscript.lua
	delLuaScript string
	delScript    = v9rds.NewScript(delLuaScript)
)

type (
	ILock interface {
		Acquire() (bool, error)
		AcquireCtx(ctx context.Context) (bool, error)
		Release() (bool, error)
		ReleaseCtx(ctx context.Context) (bool, error)
		SetExpire(seconds int)
	}

	lock struct {
		rds     *Redis
		seconds uint32
		key     string
		id      string
	}
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func NewLock(rds *Redis, key string) ILock {
	return &lock{
		rds: rds,
		key: key,
		id:  stringx.Randn(randomLen),
	}
}

func (l *lock) Acquire() (bool, error) {
	return l.AcquireCtx(context.Background())
}

func (l *lock) AcquireCtx(ctx context.Context) (bool, error) {
	seconds := atomic.LoadUint32(&l.seconds)
	resp, err := lockScript.Run(ctx, l.rds, []string{l.key}, []string{l.id, strconv.Itoa(int(seconds)*millisPerSecond + tolerance)}).Result()

	if errors.Is(err, v9rds.Nil) {
		return false, nil
	} else if err != nil {
		logx.WithContext(ctx).Errorf("Error on acquiring lock for %s, %s", l.key, err.Error())
		return false, err
	} else if resp == nil {
		return false, nil
	}

	reply, ok := resp.(string)
	if ok && reply == "OK" {
		return true, nil
	}

	logx.WithContext(ctx).Errorf("Unknown reply when acquiring lock for %s: %v", l.key, resp)
	return false, nil
}

func (l *lock) Release() (bool, error) {
	return l.ReleaseCtx(context.Background())
}

func (l *lock) ReleaseCtx(ctx context.Context) (bool, error) {
	resp, err := delScript.Run(ctx, l.rds, []string{l.key}, []string{l.id}).Result()
	if err != nil {
		return false, err
	}

	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return reply == 1, nil
}

func (l *lock) SetExpire(seconds int) {
	atomic.StoreUint32(&l.seconds, uint32(seconds))
}

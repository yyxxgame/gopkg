//@File     cache.go
//@Time     2025/5/28
//@Author   #Suyghur,

package redis

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/datetime"
	v9rds "github.com/redis/go-redis/v9"
	"github.com/yyxxgame/gopkg/stores/redis/internal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mathx"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/syncx"
)

const (
	notFoundPlaceholder = "*"
	// make the expiry unstable to avoid lots of cached items expire at the same time
	// make the unstable expiry to be [0.95, 1.05] * seconds
	expiryDeviation       = 0.05
	defaultExpiry         = time.Hour * 24
	defaultNotFoundExpiry = time.Minute
)

var (
	singleFlights  = syncx.NewSingleFlight()
	cacheStat      = internal.NewCacheStat("cache")
	errPlaceholder = errors.New("placeholder")
)

type (
	ICache interface {
		// Get gets the cache with key and fills into v.
		Get(key string, val any) error
		// GetCtx gets the cache with key and fills into v.
		GetCtx(ctx context.Context, key string, val any) error
		// Set sets the cache with key and v, using c.expiry.
		Set(key string, val any) error
		// SetCtx sets the cache with key and v, using c.expiry.
		SetCtx(ctx context.Context, key string, val any) error
		// SetWithExpire sets the cache with key and v, using given expire.
		SetWithExpire(key string, val any, expire time.Duration) error
		// SetWithExpireCtx sets the cache with key and v, using given expire.
		SetWithExpireCtx(ctx context.Context, key string, val any, expire time.Duration) error
		// Take takes the result from cache first, if not found,
		// query from DB and set cache using c.expiry, then return the result.
		Take(val any, key string, query func(val any) error) error
		// TakeCtx takes the result from cache first, if not found,
		// query from DB and set cache using c.expiry, then return the result.
		TakeCtx(ctx context.Context, val any, key string, query func(val any) error) error
		// Del deletes cached values with keys.
		Del(keys ...string) error
		// DelCtx deletes cached values with keys.
		DelCtx(ctx context.Context, keys ...string) error
	}

	cache struct {
		rds            *Redis
		expiry         time.Duration
		notFoundExpiry time.Duration
		barrier        syncx.SingleFlight
		r              *rand.Rand
		lock           *sync.Mutex
		unstableExpiry mathx.Unstable
		//stat           *Stat
		errNotFound error
	}

	CacheOption func(c *cache)
)

func NewCache(rds *Redis, opts ...CacheOption) ICache {
	c := &cache{
		rds:            rds,
		expiry:         defaultExpiry,
		notFoundExpiry: defaultNotFoundExpiry,
		barrier:        singleFlights,
		r:              rand.New(rand.NewSource(datetime.TimestampNano())),
		lock:           new(sync.Mutex),
		unstableExpiry: mathx.NewUnstable(expiryDeviation),
		errNotFound:    sql.ErrNoRows,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get gets the cache with key and fills into v.
func (c *cache) Get(key string, val any) error {
	return c.GetCtx(context.Background(), key, val)
}

// GetCtx gets the cache with key and fills into v.
func (c *cache) GetCtx(ctx context.Context, key string, val any) error {
	err := c.doGetCache(ctx, key, val)
	if errors.Is(err, errPlaceholder) {
		return c.errNotFound
	}

	return err
}

func (c *cache) Set(key string, val any) error {
	return c.SetCtx(context.Background(), key, val)
}

func (c *cache) SetCtx(ctx context.Context, key string, val any) error {
	return c.SetWithExpireCtx(ctx, key, val, c.unstableExpiry.AroundDuration(c.expiry))
}

func (c *cache) SetWithExpire(key string, val any, expire time.Duration) error {
	return c.SetWithExpireCtx(context.Background(), key, val, expire)
}

func (c *cache) SetWithExpireCtx(ctx context.Context, key string, val any, expire time.Duration) error {
	data, err := sonic.MarshalString(val)
	if err != nil {
		return err
	}

	return c.rds.Set(ctx, key, data, expire).Err()
}

func (c *cache) Take(val any, key string, query func(val any) error) error {
	return c.TakeCtx(context.Background(), val, key, query)
}

func (c *cache) TakeCtx(ctx context.Context, val any, key string, query func(val any) error) error {
	return c.doTake(ctx, val, key, query, func(v any) error {
		return c.SetCtx(ctx, key, v)
	})
}

// Del deletes cached values with keys.
func (c *cache) Del(keys ...string) error {
	return c.DelCtx(context.Background(), keys...)
}

// DelCtx deletes cached values with keys.
func (c *cache) DelCtx(ctx context.Context, keys ...string) error {
	if len(keys) <= 0 {
		return nil
	}

	err := c.rds.Del(ctx, keys...).Err()
	if err != nil {
		logx.WithContext(ctx).Errorf("failed to clear cache with keys: %q, error: %v", strings.Join(keys, ","), err)
		c.asyncRetryDelCache(keys...)
	}

	return nil
}

func (c *cache) doGetCache(ctx context.Context, key string, v any) error {
	cacheStat.IncrementTotal()
	data, err := c.rds.Get(ctx, key).Result()
	if err != nil {
		cacheStat.IncrementMiss()

		if errors.Is(err, v9rds.Nil) {
			return c.errNotFound
		}

		return err
	}

	if len(data) == 0 {
		cacheStat.IncrementMiss()
		return c.errNotFound
	}

	cacheStat.IncrementHit()
	if data == notFoundPlaceholder {
		return errPlaceholder
	}

	return c.processCache(ctx, key, data, v)
}

func (c *cache) doTake(ctx context.Context, v any, key string, query func(v any) error, cacheVal func(v any) error) error {
	val, fresh, err := c.barrier.DoEx(key, func() (any, error) {
		if err := c.doGetCache(ctx, key, v); err != nil {
			if errors.Is(err, errPlaceholder) {
				return nil, c.errNotFound
			} else if !errors.Is(err, c.errNotFound) {
				// why we just return the error instead of query from db,
				// because we don't allow the disaster pass to the dbs.
				// fail fast, in case we bring down the dbs.
				return nil, err
			}

			if err = query(v); errors.Is(err, c.errNotFound) {
				if err = c.setCacheWithNotFound(ctx, key); err != nil {
					logx.WithContext(ctx).Error(err)
				}

				return nil, c.errNotFound
			} else if err != nil {
				cacheStat.IncrementDbFails()
				return nil, err
			}

			if err = cacheVal(v); err != nil {
				logx.WithContext(ctx).Error(err)
			}
		}

		return sonic.MarshalString(v)
	})

	if err != nil {
		return err
	}
	if fresh {
		return nil
	}

	// got the result from previous ongoing query.
	// why not call IncrementTotal at the beginning of this function?
	// because a shared error is returned, and we don't want to count.
	// for example, if the db is down, the query will be failed, we count
	// the shared errors with one db failure.
	cacheStat.IncrementTotal()
	cacheStat.IncrementHit()

	return sonic.UnmarshalString(val.(string), v)
}

func (c *cache) processCache(ctx context.Context, key, data string, v any) error {
	err := sonic.UnmarshalString(data, v)
	if err == nil {
		return nil
	}

	report := fmt.Sprintf("%s unmarshal cache key: %s, value: %s, error: %v", c.rds.String(), key, data, err)
	logx.WithContext(ctx).Error(report)
	stat.Report(report)
	if e := c.rds.Del(ctx, key).Err(); e != nil {
		logx.WithContext(ctx).Errorf("%s delete invalid cache key: %s, value: %s, error: %v", c.rds.String(), key, data, e)
	}

	// returns errNotFound to reload the value by the given queryFn
	return c.errNotFound
}

func (c *cache) setCacheWithNotFound(ctx context.Context, key string) error {
	return c.rds.SetNX(ctx, key, notFoundPlaceholder, c.notFoundExpiry).Err()
}

func (c *cache) asyncRetryDelCache(keys ...string) {
	AddCleanCacheTask(func() error {
		return c.rds.Del(context.Background(), keys...).Err()
	}, keys...)
}

// WithExpiry returns a func to customize an Options with given expiry.
func WithExpiry(expiry time.Duration) CacheOption {
	return func(c *cache) {
		c.expiry = expiry
	}
}

// WithNotFoundExpiry returns a func to customize an Options with given not found expiry.
func WithNotFoundExpiry(expiry time.Duration) CacheOption {
	return func(c *cache) {
		c.notFoundExpiry = expiry
	}
}

//@File     cached_gorm.go
//@Time     2023/09/11
//@Author   #Suyghur,

package gormc

import (
	"context"
	"database/sql"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/mathx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"time"
)

const (
	// see doc/sql-cache.md
	cacheSafeGapBetweenIndexAndPrimary = time.Second * 5
	// spanName is used to identify the span name for the SQL execution.
	spanName = "gormc"

	// make the expiry unstable to avoid lots of cached items expire at the same time
	// make the unstable expiry to be [0.95, 1.05] * seconds
	expiryDeviation = 0.05
	CachePrefix     = "gormc:cache:"
)

var (
	// ErrNotFound is an alias of gorm.ErrRecordNotFound.
	ErrNotFound = gorm.ErrRecordNotFound

	// can't use one SingleFlight per conn, because multiple conns may share the same cache key.
	singleFlights     = syncx.NewSingleFlight()
	stats             = cache.NewStat("gormc")
	gormcAttributeKey = attribute.Key("gormc.method")
)

type (

	// ExecCtxFn defines the sql exec method.
	ExecCtxFn func(conn *gorm.DB) error
	// IndexQueryCtxFn defines the query method that based on unique indexes.
	IndexQueryCtxFn func(conn *gorm.DB, v interface{}) (interface{}, error)
	// PrimaryQueryCtxFn defines the query method that based on primary keys.
	PrimaryQueryCtxFn func(conn *gorm.DB, v, primary interface{}) error
	// QueryCtxFn defines the query method.
	QueryCtxFn func(conn *gorm.DB, v interface{}) error

	IGormc interface {
		SetCache(key string, value interface{}) error
		SetCacheCtx(ctx context.Context, key string, value interface{}) error
		SetCacheWithExpire(key string, value interface{}, expire time.Duration) error
		SetCacheWithExpireCtx(ctx context.Context, key string, value interface{}, expire time.Duration) error
		GetCache(key string, value interface{}) error
		GetCacheCtx(ctx context.Context, key string, value interface{}) error
		DelCache(keys ...string) error
		DelCacheCtx(ctx context.Context, keys ...string) error

		Query(value interface{}, key string, query QueryCtxFn) error
		QueryCtx(ctx context.Context, value interface{}, key string, query QueryCtxFn) error
		doQueryCtx(ctx context.Context, value interface{}, key string, query QueryCtxFn) error
		QueryNoCache(value interface{}, query QueryCtxFn) error
		QueryNoCacheCtx(ctx context.Context, value interface{}, query QueryCtxFn) error
		doQueryNoCacheCtx(ctx context.Context, value interface{}, query QueryCtxFn) error
		QueryWithExpire(value interface{}, key string, expire time.Duration, query QueryCtxFn) error
		QueryWithExpireCtx(ctx context.Context, value interface{}, key string, expire time.Duration, query QueryCtxFn) error
		doQueryWithExpireCtx(ctx context.Context, value interface{}, key string, expire time.Duration, query QueryCtxFn) error
		QueryRowIndex(value interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error
		QueryRowIndexCtx(ctx context.Context, value interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error
		doQueryRowIndexCtx(ctx context.Context, value interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error

		Exec(exec ExecCtxFn, keys ...string) error
		ExecCtx(ctx context.Context, execCtx ExecCtxFn, keys ...string) error
		ExecNoCache(exec ExecCtxFn) error
		ExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error
		doExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error

		Transact(callback func(db *gorm.DB) error, opts ...*sql.TxOptions) error
		TransactCtx(ctx context.Context, callback func(db *gorm.DB) error, opts ...*sql.TxOptions) error
	}

	cachedConn struct {
		db                 *gorm.DB
		cache              cache.Cache
		tracer             oteltrace.Tracer
		unstableExpiryTime mathx.Unstable
	}
)

// NewConn returns a CachedConn with a redis cluster cache.
func NewConn(db *gorm.DB, c cache.CacheConf, opts ...Option) IGormc {
	o := newOptions(opts...)
	cc := cache.New(c, singleFlights, stats, ErrNotFound, cache.WithExpiry(o.Expiry), cache.WithNotFoundExpiry(o.NotFoundExpiry))
	return newConnWithCache(db, cc, o)
}

// NewNodeConn returns a CachedConn with a redis node cache.
func NewNodeConn(db *gorm.DB, rds *redis.Redis, opts ...Option) IGormc {
	o := newOptions(opts...)
	cc := cache.NewNode(rds, singleFlights, stats, ErrNotFound, cache.WithExpiry(o.Expiry), cache.WithNotFoundExpiry(o.NotFoundExpiry))
	return newConnWithCache(db, cc, o)
}

// NewConnWithCache returns a CachedConn with a custom cache.
func newConnWithCache(db *gorm.DB, c cache.Cache, o Options) IGormc {
	return &cachedConn{
		db:                 db,
		cache:              c,
		tracer:             o.tracer,
		unstableExpiryTime: mathx.NewUnstable(expiryDeviation),
	}
}

// SetCache sets value into cache with given key.
func (cc *cachedConn) SetCache(key string, value interface{}) error {
	return cc.cache.SetCtx(context.Background(), key, value)
}

// SetCacheCtx sets value into cache with given key.
func (cc *cachedConn) SetCacheCtx(ctx context.Context, key string, value interface{}) error {
	return cc.cache.SetCtx(ctx, key, value)
}

// SetCacheWithExpire sets value into cache with given key.
func (cc *cachedConn) SetCacheWithExpire(key string, value interface{}, expire time.Duration) error {
	return cc.SetCacheWithExpireCtx(context.Background(), key, value, expire)
}

// SetCacheWithExpireCtx sets value into cache with given key.
func (cc *cachedConn) SetCacheWithExpireCtx(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return cc.cache.SetWithExpireCtx(ctx, key, value, expire)
}

// GetCache unmarshals cache with given key into value.
func (cc *cachedConn) GetCache(key string, value interface{}) error {
	return cc.cache.GetCtx(context.Background(), key, value)
}

// GetCacheCtx unmarshals cache with given key into value.
func (cc *cachedConn) GetCacheCtx(ctx context.Context, key string, value interface{}) error {
	return cc.cache.GetCtx(ctx, key, value)
}

// DelCache deletes cache with keys.
func (cc *cachedConn) DelCache(keys ...string) error {
	return cc.cache.DelCtx(context.Background(), keys...)
}

// DelCacheCtx deletes cache with keys.
func (cc *cachedConn) DelCacheCtx(ctx context.Context, keys ...string) error {
	return cc.cache.DelCtx(ctx, keys...)
}

func (cc *cachedConn) Query(value interface{}, key string, query QueryCtxFn) error {
	return cc.QueryCtx(context.Background(), value, key, query)
}

func (cc *cachedConn) QueryCtx(ctx context.Context, value interface{}, key string, query QueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryCtx(spanCtx, value, key, query)
		}, gormcAttributeKey.String("Query"))
	}
	return cc.doQueryCtx(ctx, value, key, query)
}

func (cc *cachedConn) doQueryCtx(ctx context.Context, value interface{}, key string, query QueryCtxFn) error {
	return cc.cache.TakeCtx(ctx, value, key, func(v interface{}) error {
		return query(cc.db.WithContext(ctx), v)
	})
}

func (cc *cachedConn) QueryNoCache(value interface{}, query QueryCtxFn) error {
	return cc.QueryNoCacheCtx(context.Background(), value, query)
}

func (cc *cachedConn) QueryNoCacheCtx(ctx context.Context, value interface{}, query QueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryNoCacheCtx(spanCtx, value, query)
		}, gormcAttributeKey.String("QueryNoCache"))
	}
	return cc.doQueryNoCacheCtx(ctx, value, query)
}

func (cc *cachedConn) doQueryNoCacheCtx(ctx context.Context, value interface{}, query QueryCtxFn) error {
	return query(cc.db.WithContext(ctx), value)
}

// QueryWithExpire unmarshals into value with given key, set expire duration and query func.
func (cc *cachedConn) QueryWithExpire(value interface{}, key string, expire time.Duration, query QueryCtxFn) error {
	return cc.QueryWithExpireCtx(context.Background(), value, key, expire, query)
}

// QueryWithExpireCtx unmarshals into value with given key, set expire duration and query func.
func (cc *cachedConn) QueryWithExpireCtx(ctx context.Context, value interface{}, key string, expire time.Duration, query QueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryWithExpireCtx(spanCtx, value, key, expire, query)
		}, gormcAttributeKey.String("QueryWithExpire"))
	}
	return cc.doQueryWithExpireCtx(ctx, value, key, expire, query)
}

func (cc *cachedConn) doQueryWithExpireCtx(ctx context.Context, value interface{}, key string, expire time.Duration, query QueryCtxFn) error {
	if err := query(cc.db.WithContext(ctx), value); err != nil {
		return err
	}
	return cc.cache.SetWithExpireCtx(ctx, key, value, cc.aroundDuration(expire))
}

// QueryRowIndex unmarshals into value with given key.
func (cc *cachedConn) QueryRowIndex(value interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	return cc.QueryRowIndexCtx(context.Background(), value, key, keyer, indexQuery, primaryQuery)
}

// QueryRowIndexCtx unmarshals into value with given key.
func (cc *cachedConn) QueryRowIndexCtx(ctx context.Context, value interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryRowIndexCtx(spanCtx, value, key, keyer, indexQuery, primaryQuery)
		}, gormcAttributeKey.String("QueryRowIndex"))
	}
	return cc.doQueryRowIndexCtx(ctx, value, key, keyer, indexQuery, primaryQuery)
}

func (cc *cachedConn) doQueryRowIndexCtx(ctx context.Context, value interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	var primaryKey interface{}
	var found bool

	if err := cc.cache.TakeWithExpireCtx(ctx, &primaryKey, key, func(val interface{}, expire time.Duration) error {
		if primaryKey, err := indexQuery(cc.db.WithContext(ctx), value); err != nil {
			return err
		} else {
			found = true
			return cc.cache.SetWithExpireCtx(ctx, keyer(primaryKey), value, expire+cacheSafeGapBetweenIndexAndPrimary)
		}
	}); err != nil {
		return err
	}
	if found {
		return nil
	}
	return cc.cache.TakeCtx(ctx, value, keyer(primaryKey), func(value interface{}) error {
		return primaryQuery(cc.db.WithContext(ctx), value, primaryKey)
	})
}

// Exec runs given exec on given keys, and returns execution result.
func (cc *cachedConn) Exec(exec ExecCtxFn, keys ...string) error {
	return cc.ExecCtx(context.Background(), exec, keys...)
}

// ExecCtx runs given exec on given keys, and returns execution result.
func (cc *cachedConn) ExecCtx(ctx context.Context, execCtx ExecCtxFn, keys ...string) error {
	err := execCtx(cc.db.WithContext(ctx))
	if err != nil {
		return err
	}
	if err := cc.DelCacheCtx(ctx, keys...); err != nil {
		return err
	}
	return nil
}

// ExecNoCache runs exec with given sql statement, without affecting cache.
func (cc *cachedConn) ExecNoCache(exec ExecCtxFn) error {
	return cc.ExecNoCacheCtx(context.Background(), exec)
}

// ExecNoCacheCtx runs exec with given sql statement, without affecting cache.
func (cc *cachedConn) ExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doExecNoCacheCtx(spanCtx, execCtx)
		}, gormcAttributeKey.String("ExecNoCache"))
	}
	return cc.doExecNoCacheCtx(ctx, execCtx)
}

func (cc *cachedConn) doExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error {
	return execCtx(cc.db.WithContext(ctx))
}

// Transact runs given fn in transaction mode.
func (cc *cachedConn) Transact(callback func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.TransactCtx(context.Background(), callback, opts...)
}

// TransactCtx runs given fn in transaction mode.
func (cc *cachedConn) TransactCtx(ctx context.Context, callback func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.db.WithContext(ctx).Transaction(callback, opts...)
}

func (cc *cachedConn) aroundDuration(duration time.Duration) time.Duration {
	return cc.unstableExpiryTime.AroundDuration(duration)
}

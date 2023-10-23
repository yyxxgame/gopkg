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

	CachedConn struct {
		db                 *gorm.DB
		cache              cache.Cache
		tracer             oteltrace.Tracer
		unstableExpiryTime mathx.Unstable
	}
)

// NewConn returns a CachedConn with a redis cluster cache.
func NewConn(db *gorm.DB, c cache.CacheConf, opts ...Option) *CachedConn {
	o := newOptions(opts...)
	cc := cache.New(c, singleFlights, stats, ErrNotFound, cache.WithExpiry(o.Expiry), cache.WithNotFoundExpiry(o.NotFoundExpiry))
	return newConnWithCache(db, cc, o)
}

// NewNodeConn returns a CachedConn with a redis node cache.
func NewNodeConn(db *gorm.DB, rds *redis.Redis, opts ...Option) *CachedConn {
	o := newOptions(opts...)
	cc := cache.NewNode(rds, singleFlights, stats, ErrNotFound, cache.WithExpiry(o.Expiry), cache.WithNotFoundExpiry(o.NotFoundExpiry))
	return newConnWithCache(db, cc, o)
}

// NewConnWithCache returns a CachedConn with a custom cache.
func newConnWithCache(db *gorm.DB, c cache.Cache, o Options) *CachedConn {
	if o.enableMetric {
		_ = db.Use(newMetricPlugin())
	}
	return &CachedConn{
		db:                 db,
		cache:              c,
		tracer:             o.tracer,
		unstableExpiryTime: mathx.NewUnstable(expiryDeviation),
	}
}

// DelCache deletes cache with keys.
func (cc *CachedConn) DelCache(keys ...string) error {
	return cc.cache.DelCtx(context.Background(), keys...)
}

// DelCacheCtx deletes cache with keys.
func (cc *CachedConn) DelCacheCtx(ctx context.Context, keys ...string) error {
	return cc.cache.DelCtx(ctx, keys...)
}

// GetCache unmarshals cache with given key into v.
func (cc *CachedConn) GetCache(key string, v interface{}) error {
	return cc.cache.GetCtx(context.Background(), key, v)
}

// GetCacheCtx unmarshals cache with given key into v.
func (cc *CachedConn) GetCacheCtx(ctx context.Context, key string, v interface{}) error {
	return cc.cache.GetCtx(ctx, key, v)
}

// Exec runs given exec on given keys, and returns execution result.
func (cc *CachedConn) Exec(exec ExecCtxFn, keys ...string) error {
	return cc.ExecCtx(context.Background(), exec, keys...)
}

// ExecCtx runs given exec on given keys, and returns execution result.
func (cc *CachedConn) ExecCtx(ctx context.Context, execCtx ExecCtxFn, keys ...string) error {
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
func (cc *CachedConn) ExecNoCache(exec ExecCtxFn) error {
	return cc.ExecNoCacheCtx(context.Background(), exec)
}

// ExecNoCacheCtx runs exec with given sql statement, without affecting cache.
func (cc *CachedConn) ExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doExecNoCacheCtx(spanCtx, execCtx)
		}, gormcAttributeKey.String("ExecNoCache"))
	}
	return cc.doExecNoCacheCtx(ctx, execCtx)
}

func (cc *CachedConn) doExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) error {
	return execCtx(cc.db.WithContext(ctx))
}

// QueryRowIndex unmarshals into v with given key.
func (cc *CachedConn) QueryRowIndex(v interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	return cc.QueryRowIndexCtx(context.Background(), v, key, keyer, indexQuery, primaryQuery)
}

// QueryRowIndexCtx unmarshals into v with given key.
func (cc *CachedConn) QueryRowIndexCtx(ctx context.Context, v interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryRowIndexCtx(spanCtx, v, key, keyer, indexQuery, primaryQuery)
		}, gormcAttributeKey.String("QueryRowIndex"))
	}
	return cc.doQueryRowIndexCtx(ctx, v, key, keyer, indexQuery, primaryQuery)
}

func (cc *CachedConn) doQueryRowIndexCtx(ctx context.Context, v interface{}, key string, keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	var primaryKey interface{}
	var found bool

	if err := cc.cache.TakeWithExpireCtx(ctx, &primaryKey, key, func(val interface{}, expire time.Duration) error {
		if primaryKey, err := indexQuery(cc.db.WithContext(ctx), v); err != nil {
			return err
		} else {
			found = true
			return cc.cache.SetWithExpireCtx(ctx, keyer(primaryKey), v, expire+cacheSafeGapBetweenIndexAndPrimary)
		}
	}); err != nil {
		return err
	}
	if found {
		return nil
	}
	return cc.cache.TakeCtx(ctx, v, keyer(primaryKey), func(v interface{}) error {
		return primaryQuery(cc.db.WithContext(ctx), v, primaryKey)
	})
}

func (cc *CachedConn) Query(v interface{}, key string, query QueryCtxFn) error {
	return cc.QueryCtx(context.Background(), v, key, query)
}

func (cc *CachedConn) QueryCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryCtx(spanCtx, v, key, query)
		}, gormcAttributeKey.String("Query"))
	}
	return cc.doQueryCtx(ctx, v, key, query)
}

func (cc *CachedConn) doQueryCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) error {
	return cc.cache.TakeCtx(ctx, v, key, func(v interface{}) error {
		return query(cc.db.WithContext(ctx), v)
	})
}

func (cc *CachedConn) QueryNoCache(v interface{}, query QueryCtxFn) error {
	return cc.QueryNoCacheCtx(context.Background(), v, query)
}

func (cc *CachedConn) QueryNoCacheCtx(ctx context.Context, v interface{}, query QueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryNoCacheCtx(spanCtx, v, query)
		}, gormcAttributeKey.String("QueryNoCache"))
	}
	return cc.doQueryNoCacheCtx(ctx, v, query)
}

func (cc *CachedConn) doQueryNoCacheCtx(ctx context.Context, v interface{}, query QueryCtxFn) error {
	return query(cc.db.WithContext(ctx), v)
}

// QueryWithExpire unmarshals into v with given key, set expire duration and query func.
func (cc *CachedConn) QueryWithExpire(v interface{}, key string, expire time.Duration, query QueryCtxFn) error {
	return cc.QueryWithExpireCtx(context.Background(), v, key, expire, query)
}

// QueryWithExpireCtx unmarshals into v with given key, set expire duration and query func.
func (cc *CachedConn) QueryWithExpireCtx(ctx context.Context, v interface{}, key string, expire time.Duration, query QueryCtxFn) error {
	if cc.tracer != nil {
		return xtrace.WithTraceHook(ctx, cc.tracer, oteltrace.SpanKindClient, spanName, func(spanCtx context.Context) error {
			return cc.doQueryWithExpireCtx(spanCtx, v, key, expire, query)
		}, gormcAttributeKey.String("QueryWithExpire"))
	}
	return cc.doQueryWithExpireCtx(ctx, v, key, expire, query)
}

func (cc *CachedConn) doQueryWithExpireCtx(ctx context.Context, v interface{}, key string, expire time.Duration, query QueryCtxFn) error {
	if err := query(cc.db.WithContext(ctx), v); err != nil {
		return err
	}
	return cc.cache.SetWithExpireCtx(ctx, key, v, cc.aroundDuration(expire))
}

func (cc *CachedConn) aroundDuration(duration time.Duration) time.Duration {
	return cc.unstableExpiryTime.AroundDuration(duration)
}

// SetCache sets v into cache with given key.
func (cc *CachedConn) SetCache(key string, v interface{}) error {
	return cc.cache.SetCtx(context.Background(), key, v)
}

// SetCacheCtx sets v into cache with given key.
func (cc *CachedConn) SetCacheCtx(ctx context.Context, key string, val interface{}) error {
	return cc.cache.SetCtx(ctx, key, val)
}

// SetCacheWithExpireCtx sets v into cache with given key.
func (cc *CachedConn) SetCacheWithExpireCtx(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	return cc.cache.SetWithExpireCtx(ctx, key, val, expire)
}

// Transact runs given fn in transaction mode.
func (cc *CachedConn) Transact(fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.TransactCtx(context.Background(), fn, opts...)
}

// TransactCtx runs given fn in transaction mode.
func (cc *CachedConn) TransactCtx(ctx context.Context, fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.db.WithContext(ctx).Transaction(fn, opts...)
}

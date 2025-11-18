//@File     redis.go
//@Time     2025/5/28
//@Author   #Suyghur,

package redis

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	v9rds "github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/yyxxgame/gopkg/stores/redis/internal"
	"github.com/zeromicro/go-zero/core/syncx"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
)

const (
	maxRetries            = 3
	idleConns             = 8
	slowThresholdDuration = time.Millisecond * 100
	readTimeout           = 2 * time.Second
	pingTimeout           = time.Second
)

type (
	Redis struct {
		*v9rds.Client
		db                    int
		maxRetries            int
		idleConns             int
		slowThresholdDuration time.Duration
		readTimeout           time.Duration
		pingTimeout           time.Duration
		hooks                 []v9rds.Hook
		// 适配redis 7.x以下版本
		disableIdentity          bool
		maintNotificationsConfig *maintnotifications.Config
	}

	Option func(rds *Redis)
)

func NewRedis(conf RedisConf, opts ...Option) *Redis {
	rds := &Redis{
		db:                    conf.DB,
		maxRetries:            maxRetries,
		idleConns:             idleConns,
		slowThresholdDuration: slowThresholdDuration,
		readTimeout:           readTimeout,
		pingTimeout:           pingTimeout,
		disableIdentity:       false,
		maintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeAuto,
		},
	}

	for _, opt := range opts {
		opt(rds)
	}

	rds.Client = v9rds.NewClient(&v9rds.Options{
		Addr:                     conf.Host,
		Password:                 conf.Pass,
		DB:                       rds.db,
		MaxRetries:               rds.maxRetries,
		MinIdleConns:             rds.idleConns,
		ReadTimeout:              rds.readTimeout,
		DisableIdentity:          rds.disableIdentity,
		MaintNotificationsConfig: rds.maintNotificationsConfig,
	})

	rds.Client.AddHook(&internal.DurationHook{
		Tracer:        otel.GetTracerProvider().Tracer(gozerotrace.TraceName),
		SlowThreshold: syncx.ForAtomicDuration(rds.slowThresholdDuration),
	})

	for _, hook := range rds.hooks {
		rds.Client.AddHook(hook)
	}

	internal.ConnCollector.RegisterClient(&internal.StatGetter{
		Host:     conf.Host,
		DB:       rds.db,
		PoolSize: 10 * runtime.GOMAXPROCS(0),
		PoolStats: func() *v9rds.PoolStats {
			return rds.Client.PoolStats()
		},
	})

	if err := rds.checkConnection(rds.pingTimeout); err != nil {
		panic(errors.New(fmt.Sprintf("redis connect error, addr: %s", conf.Host)))
		return nil
	}

	return rds
}

func (rds *Redis) checkConnection(pingTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	return rds.Client.Ping(ctx).Err()
}

func (rds *Redis) ScanKeys(ctx context.Context, pattern string, count int64) []string {
	var keys []string
	iter := rds.Scan(ctx, 0, pattern, count).Iterator()
	for iter.Next(ctx) {
		if iter.Err() != nil {
			break
		}
		keys = append(keys, iter.Val())
	}

	if len(keys) <= 0 {
		keys = make([]string, 0)
	}

	return keys
}

func WithDB(db int) Option {
	return func(rds *Redis) {
		rds.db = db
	}
}

func WithIdleConns(conns int) Option {
	return func(rds *Redis) {
		rds.idleConns = conns
	}
}

func WithMaxRetries(retries int) Option {
	return func(rds *Redis) {
		rds.maxRetries = retries
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(rds *Redis) {
		rds.readTimeout = timeout
	}
}

func WithPingTimeout(timeout time.Duration) Option {
	return func(rds *Redis) {
		rds.pingTimeout = timeout
	}
}

func WithSlowThresholdDuration(threshold time.Duration) Option {
	return func(rds *Redis) {
		rds.slowThresholdDuration = threshold
	}
}

func WithDisableIdentity() Option {
	return func(rds *Redis) {
		rds.disableIdentity = true
	}
}

func WithMaintNotificationsConfig(conf *maintnotifications.Config) Option {
	return func(rds *Redis) {
		rds.maintNotificationsConfig = conf
	}
}

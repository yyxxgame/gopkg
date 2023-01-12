//@File     client.go
//@Time     2023/01/05
//@Author   #Suyghur,

package etcdv3

import (
	"context"
	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc"
	"time"
)

type (
	Option func(impl *client)
	client struct {
		instance             *clientv3.Client
		enableAuth           bool
		username             string
		password             string
		dialTimeout          time.Duration
		dialKeepAliveTime    time.Duration
		dialKeepAliveTimeout time.Duration
	}
	IEtcdv3 interface {
		GetKeyValue(key string) (*mvccpb.KeyValue, error)
		GetKeyValueCtx(ctx context.Context, key string) (*mvccpb.KeyValue, error)
		GetPrefix(prefix string) (map[string]string, error)
		GetPrefixCtx(ctx context.Context, prefix string) (map[string]string, error)
		DelPrefix(prefix string) (int64, error)
		DelPrefixCtx(ctx context.Context, prefix string) (int64, error)
		GetValues(keys ...string) (map[string]string, error)
		GetValuesCtx(ctx context.Context, keys ...string) (map[string]string, error)
		GetLeaseSession(opts ...concurrency.SessionOption) (*concurrency.Session, error)
		GetLeaseSessionCtx(ctx context.Context, opts ...concurrency.SessionOption) (*concurrency.Session, error)
	}
)

func New(endpoints []string, opts ...Option) IEtcdv3 {
	impl := &client{}
	for _, opt := range opts {
		opt(impl)
	}

	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(grpcprom.StreamClientInterceptor),
	}

	if impl.dialTimeout == 0 {
		impl.dialTimeout = 5 * time.Second
	}
	if impl.dialKeepAliveTime == 0 {
		impl.dialKeepAliveTime = 10 * time.Second
	}
	if impl.dialKeepAliveTimeout == 0 {
		impl.dialKeepAliveTimeout = 3 * time.Second
	}

	conf := clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          impl.dialTimeout,
		DialKeepAliveTime:    impl.dialKeepAliveTime,
		DialKeepAliveTimeout: impl.dialKeepAliveTimeout,
		DialOptions:          dialOptions,
	}

	if impl.enableAuth {
		conf.Username = impl.username
		conf.Password = impl.password
	}

	if cli, err := clientv3.New(conf); err != nil {
		logx.WithContext(context.Background()).Error(err.Error())
		panic(err.Error())
	} else {
		impl.instance = cli
	}
	return impl
}

func WithAuth(username, password string) Option {
	return func(impl *client) {
		impl.enableAuth = true
		impl.username = username
		impl.password = password
	}
}

func WithDialTimeout(t time.Duration) Option {
	return func(impl *client) {
		impl.dialTimeout = t
	}
}
func WithDialKeepAliveTime(t time.Duration) Option {
	return func(impl *client) {
		impl.dialKeepAliveTime = t
	}
}

func WithDialKeepAliveTimeout(t time.Duration) Option {
	return func(impl *client) {
		impl.dialKeepAliveTimeout = t
	}
}

func (impl *client) GetKeyValue(key string) (*mvccpb.KeyValue, error) {
	return impl.GetKeyValueCtx(context.Background(), key)
}

func (impl *client) GetKeyValueCtx(ctx context.Context, key string) (*mvccpb.KeyValue, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *client) GetPrefix(prefix string) (map[string]string, error) {
	return impl.GetPrefixCtx(context.Background(), prefix)
}

func (impl *client) GetPrefixCtx(ctx context.Context, prefix string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *client) DelPrefix(prefix string) (int64, error) {
	return impl.DelPrefixCtx(context.Background(), prefix)
}

func (impl *client) DelPrefixCtx(ctx context.Context, prefix string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *client) GetValues(keys ...string) (map[string]string, error) {
	return impl.GetValuesCtx(context.Background(), keys...)
}

func (impl *client) GetValuesCtx(ctx context.Context, keys ...string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *client) GetLeaseSession(opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	return impl.GetLeaseSessionCtx(context.Background(), opts...)
}

func (impl *client) GetLeaseSessionCtx(ctx context.Context, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	//TODO implement me
	panic("implement me")
}

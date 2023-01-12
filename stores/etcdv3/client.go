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
	"strings"
	"time"
)

type (
	Option func(impl *client)
	client struct {
		*clientv3.Client
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
		impl.Client = cli
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

// GetKeyValue queries etcd key, returns mvccpb.KeyValue
func (impl *client) GetKeyValue(key string) (*mvccpb.KeyValue, error) {
	return impl.GetKeyValueCtx(context.Background(), key)
}

// GetKeyValueCtx queries etcd key, returns mvccpb.KeyValue
func (impl *client) GetKeyValueCtx(ctx context.Context, key string) (*mvccpb.KeyValue, error) {
	if resp, err := impl.Get(ctx, key); err != nil {
		return nil, err
	} else {
		if len(resp.Kvs) > 0 {
			return resp.Kvs[0], nil
		} else {
			return nil, nil
		}
	}
}

// GetPrefix get by prefix
func (impl *client) GetPrefix(prefix string) (map[string]string, error) {
	return impl.GetPrefixCtx(context.Background(), prefix)
}

// GetPrefixCtx get by prefix
func (impl *client) GetPrefixCtx(ctx context.Context, prefix string) (map[string]string, error) {
	result := make(map[string]string)
	if resp, err := impl.Get(ctx, prefix, clientv3.WithPrefix()); err != nil {
		return result, err
	} else {
		for _, kv := range resp.Kvs {
			result[string(kv.Key)] = string(kv.Value)
		}
	}
	return result, nil
}

// DelPrefix delete by prefix
func (impl *client) DelPrefix(prefix string) (int64, error) {
	return impl.DelPrefixCtx(context.Background(), prefix)
}

// DelPrefixCtx delete by prefix
func (impl *client) DelPrefixCtx(ctx context.Context, prefix string) (int64, error) {
	if resp, err := impl.Delete(ctx, prefix, clientv3.WithPrefix()); err != nil {
		return 0, err
	} else {
		return resp.Deleted, err
	}
}

// GetValues queries etcd for keys prefixed by prefix.
func (impl *client) GetValues(keys ...string) (map[string]string, error) {
	return impl.GetValuesCtx(context.Background(), keys...)
}

// GetValuesCtx queries etcd for keys prefixed by prefix.
func (impl *client) GetValuesCtx(ctx context.Context, keys ...string) (map[string]string, error) {
	firstRevision := int64(0)
	result := make(map[string]string)
	maxTxnOps := 128
	doTxn := func(ops []string) error {
		txnOps := make([]clientv3.Op, 0, maxTxnOps)
		for _, k := range ops {
			txnOps = append(txnOps, clientv3.OpGet(k, clientv3.WithPrefix(),
				clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
				clientv3.WithRev(firstRevision)))
		}
		if txnResult, err := impl.Txn(ctx).Then(txnOps...).Commit(); err != nil {
			return err
		} else {
			for i, r := range txnResult.Responses {
				originKey := ops[i]
				originKeyFixed := originKey
				if !strings.HasSuffix(originKeyFixed, "/") {
					originKeyFixed = originKey + "/"
				}
				for _, ev := range r.GetResponseRange().Kvs {
					k := string(ev.Key)
					if k == originKey || strings.HasPrefix(k, originKeyFixed) {
						result[string(ev.Key)] = string(ev.Value)
					}
				}
			}
			if firstRevision == 0 {
				firstRevision = txnResult.Header.GetRevision()
			}
			return nil
		}
	}
	cnt := len(keys) / maxTxnOps
	for i := 0; i <= cnt; i++ {
		switch temp := i == cnt; temp {
		case false:
			if err := doTxn(keys[i*maxTxnOps : (i+1)*maxTxnOps]); err != nil {
				return result, err
			}
		case true:
			if err := doTxn(keys[i*maxTxnOps:]); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

// GetLeaseSession create lease
func (impl *client) GetLeaseSession(opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	return impl.GetLeaseSessionCtx(context.Background(), opts...)
}

// GetLeaseSessionCtx create lease
func (impl *client) GetLeaseSessionCtx(ctx context.Context, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	return concurrency.NewSession(impl.Client, opts...)
}

//@File     observer.go
//@Time     2023/01/21
//@Author   #Suyghur,

package obkv

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type (
	IObserver interface {
		Attach(key string, callback func(data *Data)) *Observer
		AttachWithPrefix(key string, callback func(data *Data)) *Observer
		Detach() error
	}
	Observer struct {
		ctx context.Context
		*clientv3.Client
		version map[string]int64
		wg      *threading.RoutineGroup
		obOpts  []clientv3.OpOption
	}
)

func NewObkv(ctx context.Context, client *clientv3.Client) *Observer {
	ob := &Observer{
		ctx:     ctx,
		Client:  client,
		version: make(map[string]int64),
		wg:      threading.NewRoutineGroup(),
		obOpts:  make([]clientv3.OpOption, 0),
	}
	return ob
}

func (ob *Observer) Attach(key string, callback func(data *Data)) *Observer {
	ob.wg.RunSafe(func() {
		ob.attach(key, false, callback)
	})
}

func (ob *Observer) AttachWithPrefix(key string, callback func(data *Data)) *Observer {
	ob.wg.RunSafe(func() {
		ob.attach(key, true, callback)
	})
	return ob
}

func (ob *Observer) attach(key string, enablePrefix bool, callback func(data *Data)) {
	var watchChan clientv3.WatchChan
	if enablePrefix {
		watchChan = ob.Watch(ob.ctx, key, clientv3.WithPrefix())
	} else {
		watchChan = ob.Watch(ob.ctx, key)
	}
	for ch := range watchChan {
		if ch.Err() != nil {
			logx.WithContext(ob.ctx).Errorf("etcdv3 observer attach onError: %v", ch.Err())
			return
		}
		for _, event := range ch.Events {
			data := newData(event)
			if data.ModRevision == ob.version[data.Key] {
				continue
			}
			ob.version[data.Key] = data.ModRevision
			callback(data)
		}
	}
}

func (ob *Observer) Detach() error {
	ob.Close()
}

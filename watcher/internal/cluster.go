//@Auther   : Kaishin
//@Time     : 2022/12/22
//@Describe : etcd管理集群

package internal

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/lang"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/threading"
	client3 "go.etcd.io/etcd/client/v3"
	"io"
	"sort"
	"strings"
	"sync"
)

var (
	connManager = syncx.NewResourceManager()
)

type (
	// etcd集群实例类
	cluster struct {
		key        string
		endpoints  []string
		listeners  map[string][]EventKV
		watchGroup *threading.RoutineGroup
		done       chan lang.PlaceholderType
		lock       sync.Mutex
	}
)

func newCluster(endpoints []string) *cluster {
	return &cluster{
		endpoints: endpoints,
		key:       getClusterKey(endpoints),
		//values:     make(map[string]map[string]string),
		listeners:  make(map[string][]EventKV),
		watchGroup: threading.NewRoutineGroup(),
		//done:       make(chan lang.PlaceholderType),
	}
}

func getClusterKey(endpoints []string) string {
	sort.Strings(endpoints)
	return strings.Join(endpoints, endpointsSeparator)
}

func (sel *cluster) getClient() (*client3.Client, error) {
	val, err := connManager.GetResource(sel.key, func() (io.Closer, error) {
		return sel.newClient()
	})
	if err != nil {
		return nil, err
	}

	return val.(*client3.Client), nil
}

func (sel *cluster) newClient() (*client3.Client, error) {
	cfg := client3.Config{
		Endpoints:            sel.endpoints,
		AutoSyncInterval:     autoSyncInterval,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    dialKeepAliveTime,
		DialKeepAliveTimeout: dialTimeout,
		RejectOldCluster:     true,
	}

	return client3.New(cfg)
}

func (sel *cluster) monitor(key string, l EventKV) error {
	sel.lock.Lock()
	sel.listeners[key] = append(sel.listeners[key], l)
	sel.lock.Unlock()

	cli, err := sel.getClient()
	if err != nil {
		return err
	}

	sel.watchGroup.Run(func() {
		sel.watch(cli, key)
	})
	return nil
}

func (sel *cluster) watch(cli *client3.Client, key string) {
	for {
		if sel.watchStream(cli, key) {
			return
		}
	}
}

func (sel *cluster) watchStream(cli *client3.Client, key string) bool {
	rch := cli.Watch(client3.WithRequireLeader(cli.Ctx()), key, client3.WithPrefix())
	for {
		select {
		case resp, ok := <-rch:
			if !ok {
				logx.Error("etcd monitor chan has been closed")
				return false
			}
			if resp.Canceled {
				logx.Errorf("etcd monitor chan has been canceled, error: %v", resp.Err())
				return false
			}
			if resp.Err() != nil {
				logx.Error(fmt.Sprintf("etcd monitor chan error: %v", resp.Err()))
				return false
			}

			sel.handleWatchEvents(key, resp.Events)
		case <-sel.done:
			return true
		}
	}
}

func (sel *cluster) handleWatchEvents(key string, events []*client3.Event) {
	sel.lock.Lock()
	listeners := append([]EventKV(nil), sel.listeners[key]...)
	sel.lock.Unlock()

	for _, ev := range events {
		switch ev.Type {
		case client3.EventTypePut:
			for _, l := range listeners {
				l.OnAdd(KV{
					K: string(ev.Kv.Key),
					V: string(ev.Kv.Value),
				})
			}
		case client3.EventTypeDelete:
			for _, l := range listeners {
				l.OnDelete(KV{
					K: string(ev.Kv.Key),
					V: string(ev.Kv.Value),
				})
			}
		default:
			logx.Errorf("Unknown event type: %v", ev.Type)
		}
	}
}

//@Auther   : Kaishin
//@Time     : 2022/12/20
//@Describe : etcd监听分配器

package internal

import (
	"github.com/zeromicro/go-zero/core/logx"
	client3 "go.etcd.io/etcd/client/v3"
	"sync"
)

var (
	allocator = Allocator{
		clusters: make(map[string]*cluster),
	}
)

type (
	// Allocator 分配器
	Allocator struct {
		clusters map[string]*cluster
		lock     sync.Mutex
	}
)

func GetAllocator() *Allocator {
	return &allocator
}

func (sel *Allocator) Monitor(endpoints []string, key string, l EventKV) error {
	c, exists := sel.getCluster(endpoints)
	if !exists {
		logx.Infof("[Allocator.Monitor] create new cluster c: %+v", c)
	}
	return c.monitor(key, l)
}

func (sel *Allocator) GetConn(endpoints []string) (*client3.Client, error) {
	c, _ := sel.getCluster(endpoints)
	return c.getClient()
}

func (sel *Allocator) getCluster(endpoints []string) (c *cluster, exists bool) {
	clusterKey := getClusterKey(endpoints)
	sel.lock.Lock()
	defer sel.lock.Unlock()
	c, exists = sel.clusters[clusterKey]
	if !exists {
		c = newCluster(endpoints)
		sel.clusters[clusterKey] = c
	}
	return
}

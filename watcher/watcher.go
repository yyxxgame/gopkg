//@Auther   : Kaishin
//@Time     : 2022/12/20
// 基于etcd实现的事件触发监听器

package watcher

import (
	"github.com/yyxxgame/gopkg/watcher/internal"
	"sync"
)

type (
	WatchClient interface {
		AddListener(fc Listener) int // 新增一个监听者
		Notify(value string) error   // 触发事件
	}

	Listener func(k, v string)

	container struct {
		listeners []Listener
		lock      sync.Mutex
	}

	watcher struct {
		endpoints []string
		key       string
		items     *container
	}
)

func NewWatcher(endpoints []string, key string) (WatchClient, error) {
	wch := new(watcher)
	wch.endpoints = endpoints
	wch.key = key
	wch.items = new(container)

	// 分配器给每个触发的key值分配一个监听器
	alloc := internal.GetAllocator()
	if err := alloc.Monitor(endpoints, key, wch.items); err != nil {
		return nil, err
	}

	return wch, nil
}

func (sel *watcher) Notify(value string) error {
	/*
		触发事件更新
	*/
	alloc := internal.GetAllocator()
	etcdConn, err := alloc.GetConn(sel.endpoints)
	if err != nil {
		return err
	}
	_, err = etcdConn.Put(etcdConn.Ctx(), sel.key, value)
	if err != nil {
		return err
	}
	return nil
}

func (sel *watcher) AddListener(fc Listener) int {
	/*
		新增一个监听者
	*/
	sel.items.lock.Lock()
	defer sel.items.lock.Unlock()
	sel.items.listeners = append(sel.items.listeners, fc)
	return len(sel.items.listeners)
}

func (sel *container) OnAdd(kv internal.KV) {
	sel.onEvent(kv)
}

func (sel *container) OnDelete(kv internal.KV) {
	sel.onEvent(kv)
}

func (sel *container) onEvent(kv internal.KV) {
	sel.lock.Lock()
	listeners := append(([]Listener)(nil), sel.listeners...)
	sel.lock.Unlock()

	for _, listener := range listeners {
		listener(kv.K, kv.V)
	}
}

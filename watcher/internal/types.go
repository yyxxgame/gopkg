//@Auther   : Kaishin
//@Time     : 2022/12/22

package internal

type (
	// KV etcd更新key value值
	KV struct {
		K string
		V string
	}

	// EventKV 监听函数入口
	EventKV interface {
		OnAdd(kv KV)
		OnDelete(kv KV)
	}
)

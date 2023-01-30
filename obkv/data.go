//@File     data.go
//@Time     2023/01/30
//@Author   #Suyghur,

package obkv

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	Change EventType = iota
	Create
	Delete
)

type (
	EventType int
	MetaData  struct {
		// 键
		Key string
		// 值
		Value string
		// 当前版本
		Version int64
		// 创建版本
		CreateRevision int64
		// 修订版本
		ModRevision int64
		// 租约id
		Lease int64
	}
	Data struct {
		Type EventType
		*MetaData
		Old *MetaData
	}
)

func (eventType EventType) String() string {
	switch eventType {
	case Create:
		return "Create"
	case Delete:
		return "Delete"
	}
	return "Change"
}

func parseMetaDataFromRaw(kv *mvccpb.KeyValue) *MetaData {
	return &MetaData{
		Key:            string(kv.Key),
		Value:          string(kv.Value),
		Version:        kv.Version,
		CreateRevision: kv.CreateRevision,
		ModRevision:    kv.ModRevision,
		Lease:          kv.Lease,
	}
}

func newData(e *clientv3.Event) *Data {
	data := &Data{
		MetaData: parseMetaDataFromRaw(e.Kv),
	}

	if e.PrevKv != nil {
		data.Old = parseMetaDataFromRaw(e.PrevKv)
	}

	if e.Type == clientv3.EventTypeDelete {
		data.Type = Delete
	} else if data.CreateRevision == data.ModRevision {
		data.Type = Create
	} else {
		data.Type = Change
	}
	return data
}

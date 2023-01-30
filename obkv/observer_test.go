//@File     observer_test.go
//@Time     2023/01/30
//@Author   #Suyghur,

package obkv

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stringx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestNewObkv(t *testing.T) {
	key := "test_ob"
	value := stringx.Rand()
	cli, _ := clientv3.NewFromURL("http://127.0.0.1:2379")
	ctx := context.Background()
	NewObkv(ctx, cli).Attach("test_ob", func(data *Data) {
		assert.Equal(t, key, data.Key)
		assert.Equal(t, value, data.Value)
	})
	cli.Put(ctx, key, value)
	time.Sleep(5 * time.Second)
}

func TestWithPrefix(t *testing.T) {
	prefixKey := "test_"
	key1 := "test_ob1"
	key2 := "test_ob2"
	value1 := stringx.Rand()
	value2 := stringx.Rand()
	cli, _ := clientv3.NewFromURL("http://127.0.0.1:2379")
	ctx := context.Background()
	NewObkv(ctx, cli).AttachWithPrefix(prefixKey, func(data *Data) {
		//assert.Equal(t, value, data.Value)
		if data.Key != key1 && data.Key != key2 {
			t.FailNow()
		}
		if data.Value != value1 && data.Value != value2 {
			t.FailNow()
		}
	})
	cli.Put(ctx, key1, value1)
	cli.Put(ctx, key2, value2)
	time.Sleep(5 * time.Second)
}

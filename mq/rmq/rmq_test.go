//@File     rmq_test.go
//@Time     2023/07/17
//@Author   #Suyghur,

package rmq

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/stringx"
	"testing"
)

type testData struct {
	Opt     int32  `json:"opt"`
	Payload string `json:"payload"`
}

func TestRmqProducer(t *testing.T) {
	p := NewRmqProducer("loaclhost:6379", WithPassword("redis-pwd"))
	defer p.Release()
	data := &testData{
		Opt:     1000,
		Payload: stringx.Rand(),
	}
	bMsg, _ := json.Marshal(data)
	p.Publish("test_rmq", bMsg)
}

func TestRmqConsumer(t *testing.T) {
	c := NewRmqConsumer("loaclhost:6379", "test_rmq", WithPassword("redis-pwd"))
	defer c.Release()
	c.LooperSync("test_rmq", func(ctx context.Context, message *Message) error {
		t.Logf("message topic: %s", string(message.Topic))
		t.Logf("message payload: %s", string(message.Payload))
		return nil
	})
}

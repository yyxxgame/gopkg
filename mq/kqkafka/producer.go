//@File     producer.go
//@Time     2023/05/08
//@Author   LvWenQi

package kqkafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"sync"
)

type (
	Producer struct {
		pushers sync.Map
		brokers []string
	}

	ProducerInterface interface {
		GetPusher(topic string) *kq.Pusher
		Publish(ctx context.Context, topic, v string) error
	}
)

func NewProducer(brokers []string) ProducerInterface {
	logx.Info("[kqkafka.Producer.NewProducer] create, brokers:", brokers)
	return &Producer{
		brokers: brokers,
	}
}

func (sel *Producer) GetPusher(topic string) *kq.Pusher {
	obj, ok := sel.pushers.Load(topic)
	if ok == true {
		return obj.(*kq.Pusher)
	}
	newPusher := kq.NewPusher(sel.brokers, topic)
	sel.pushers.Store(topic, newPusher)
	logx.Info("[kqkafka.Producer.GetPusher] create new pusher, topic:", topic)
	return newPusher
}

func (sel *Producer) Publish(ctx context.Context, topic, v string) error {
	span, mqMsg := xtrace.StartMqProducerTrace(
		ctx,
		fmt.Sprintf("%s.%s", "kqkafka.Publish", topic),
		attribute.String("topic", topic),
	)
	defer span.End()
	mqMsg.Body = v
	msg, _ := json.Marshal(mqMsg)
	return sel.GetPusher(topic).Push(string(msg))
}

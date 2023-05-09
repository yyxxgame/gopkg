//@File     producer.go
//@Time     2023/05/08
//@Author   LvWenQi

package kqkafka

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
)

type (
	Producer struct {
		pushers sync.Map
		brokers []string
	}

	ProducerInterface interface {
		GetPusher(topic string) *Pusher
		Publish(ctx context.Context, topic, v string) error
	}
)

func NewProducer(brokers []string) ProducerInterface {
	logx.Info("[kqkafka.Producer.NewProducer] create, brokers:", brokers)
	return &Producer{
		brokers: brokers,
	}
}

func (sel *Producer) GetPusher(topic string) *Pusher {
	obj, ok := sel.pushers.Load(topic)
	if ok == true {
		return obj.(*Pusher)
	}

	newPusher := NewPusher(sel.brokers, topic)
	sel.pushers.Store(topic, newPusher)
	logx.Info("[kqkafka.Producer.GetPusher] create new pusher, topic:", topic)
	return newPusher
}

func (sel *Producer) Publish(ctx context.Context, topic, v string) error {
	pusher := sel.GetPusher(topic)
	return pusher.Push(ctx, v)
}

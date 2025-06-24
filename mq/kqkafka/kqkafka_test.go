//@File     kqkafka_test.go
//@Time     2023/05/08
//@Author   LvWenQi

package kqkafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

var (
	brokers   = []string{"localhost:9092"}
	topic     = "test_kq_kafka"
	groupId   = "test_group"
	kafkaConf = kq.KqConf{
		ServiceConf: service.ServiceConf{Name: "kqkafka_test"},
		Brokers:     brokers,
		Group:       groupId,
		Topic:       topic,
		Consumers:   1,
		Processors:  1,
		Offset:      "last",
	}
)

func TestKqKafkaProducer(t *testing.T) {
	p := NewProducer(brokers)
	_ = p.Publish(context.Background(), topic, "test_value1111111")
	_ = p.GetPusher(topic).Close()
}

func TestKqKafkaConsumer(t *testing.T) {
	consumer := NewConsumer(kafkaConf, func(ctx context.Context, key, value string) error {
		fmt.Println(value)
		fmt.Println("===========================================")
		return nil
	})
	consumer.Start()
	time.Sleep(300 * time.Second)
}

//@File     ckafka_test.go
//@Time     2023/04/28
//@Author   #Suyghur,

package ckafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"testing"
	"time"
)

var (
	brokers = []string{"localhost:9092"}
	topics  = []string{"test_ckafka"}
	groupId = "test_group"
)

func TestCkafkaProducer(t *testing.T) {
	p := NewCkafkaProducer(brokers)
	p.Publish(topics[0], "test_message", []byte("test_value"))
}

func TestCkafkaConsumer(t *testing.T) {
	exit := make(chan int)
	c := NewCkafkaConsumer(brokers, topics, groupId)

	c.Looper(func(message *kafka.Message) error {
		t.Logf("handle message, key: %s, value: %s", message.Key, message.Value)
		time.Sleep(2 * time.Second)
		return nil
	})
	defer c.Release()
	<-exit
}

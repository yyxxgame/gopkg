//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/core/syncx"
)

var (
	brokers = []string{"101.33.209.36:9092"}
	topics  = []string{"testsdk"}
	groupId = "gopkg_test_group"
)

func TestSaramaKafkaProducer(t *testing.T) {
	p := NewSaramaKafkaProducer(brokers, WithProducerInterceptor(func(message *sarama.ProducerMessage) {
		logx.Infof("OnSend Interceptor...")
		logx.Infof("message: %+v", *message)
	}))
	p.Publish(topics[0], "test_message", []byte(stringx.Randn(16)))
	p.Publish(topics[0], "test_message", []byte(stringx.Randn(16)))
	p.Publish(topics[0], "test_message", []byte(stringx.Randn(16)))
	p.Publish(topics[0], "test_message", []byte(stringx.Randn(16)))
}

func TestSaramaKafkaConsumer(t *testing.T) {
	done := syncx.NewDoneChan()
	c0 := NewSaramaConsumer(brokers, topics, groupId, WithConsumerInterceptor(func(message *sarama.ConsumerMessage) {
		logx.Infof("OnConsume Interceptor...")
		logx.Infof("message: %+v", *message)
	}))
	c0.Looper(func(ctx context.Context, message *sarama.ConsumerMessage) error {
		t.Logf("consumer0 handle message, key: %s, offset: %d, value: %s", message.Key, message.Offset, message.Value)
		time.Sleep(2 * time.Second)
		return nil
	})

	proc.AddShutdownListener(func() {
		done.Close()
	})

	<-done.Done()
}

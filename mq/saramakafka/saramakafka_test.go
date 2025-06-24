//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/core/syncx"
)

var (
	brokers = []string{"localhost:9092"}
	topics  = []string{"test-topic"}
	groupId = "gopkg_test_group"
)

func TestSaramaKafkaProducer(t *testing.T) {
	onFakeHook := false
	fakeProducerHook := func(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error {
		logx.Infof("fake")
		onFakeHook = true
		return next(ctx, message)
	}
	p := NewProducer(brokers, WithProducerHook(fakeProducerHook))
	err := p.Publish(topics[0], "test_message", stringx.Randn(16))

	assert.Nil(t, err)
	assert.True(t, onFakeHook)
}

func TestSaramaKafkaConsumer(t *testing.T) {
	done := syncx.NewDoneChan()
	var (
		onFakeHookCount uint32
		onMessageCount  uint32
	)

	fakeConsumerHook := func(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error {
		atomic.AddUint32(&onFakeHookCount, 1)
		return next(ctx, message)
	}
	handler := func(ctx context.Context, message *sarama.ConsumerMessage) error {
		logx.Infof("handle message, partition: %d, offset: %d, key: %s, value: %s", message.Partition, message.Offset, message.Key, message.Value)
		atomic.AddUint32(&onMessageCount, 1)
		return nil
	}

	c0 := NewConsumer(brokers, topics, groupId, handler, WithDisableStatLag(), WithConsumerHook(fakeConsumerHook))
	c0.Loop()

	proc.AddShutdownListener(func() {
		c0.Release()
		done.Close()
	})

	<-done.Done()

	assert.Equal(t, onFakeHookCount, onMessageCount)
}

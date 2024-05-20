//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/core/syncx"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.uber.org/atomic"
)

var (
	brokers = []string{"localhost:9092"}
	topics  = []string{"test-topic"}
	groupId = "gopkg_test_group"
)

func TestSaramaKafkaProducer(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	onFakeHook := false
	fakeProducerHook := func(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error {
		logx.Infof("fake")
		onFakeHook = true
		return next(ctx, message)
	}
	p := NewSaramaKafkaProducer(brokers, WithTracer(tracer), WithProducerHook(fakeProducerHook))
	err := p.Publish(topics[0], "test_message", stringx.Randn(16))

	assert.Nil(t, err)
	assert.True(t, onFakeHook)
}

func TestSaramaKafkaConsumer(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	onFakeHookCount := atomic.NewInt64(0)
	onMessageCount := atomic.NewInt64(0)
	fakeConsumerHook := func(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error {
		onFakeHookCount.Inc()
		return next(ctx, message)
	}
	done := syncx.NewDoneChan()
	c0 := NewSaramaKafkaConsumer(brokers, topics, groupId, WithTracer(tracer), WithDisableStatLag(), WithConsumerHook(fakeConsumerHook))
	c0.Loop(func(ctx context.Context, message *sarama.ConsumerMessage) error {
		t.Logf("consumer0 handle message, key: %s, offset: %d, value: %s", message.Key, message.Offset, message.Value)
		time.Sleep(2 * time.Second)
		onMessageCount.Inc()
		return nil
	})

	proc.AddShutdownListener(func() {
		done.Close()
	})

	<-done.Done()

	assert.Equal(t, onFakeHookCount.Load(), onMessageCount.Load())
}

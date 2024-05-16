//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/core/syncx"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
)

var (
	brokers = []string{"localhost:9092"}
	topics  = []string{"test-kafka-topic"}
	groupId = "gopkg_test_group"
)

func TestSaramaKafkaProducer(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	p := NewSaramaKafkaProducer(brokers, WithTracer(tracer))
	p.Publish(topics[0], "test_message", stringx.Randn(16))
	p.Publish(topics[0], "test_message", stringx.Randn(16))
	p.Publish(topics[0], "test_message", stringx.Randn(16))
	p.Publish(topics[0], "test_message", stringx.Randn(16))
}

func TestSaramaKafkaConsumer(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	done := syncx.NewDoneChan()
	c0 := NewSaramaKafkaConsumer(brokers, topics, groupId, WithTracer(tracer), WithEnableStatLag())
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

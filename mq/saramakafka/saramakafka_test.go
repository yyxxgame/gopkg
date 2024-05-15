//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"testing"

	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"

	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	brokers = []string{"101.33.209.36:9092"}
	topics  = []string{"testsdk"}
	groupId = "gopkg_test_group"
)

func TestSaramaKafkaProducer(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	p := NewSaramaKafkaProducer(brokers, WithTracer(tracer))
	p.Publish(topics[0], "test_message", stringx.Randn(16))
}

func TestSaramaKafkaConsumer(t *testing.T) {
	//interceptor := &fakeInterceptor{
	//	before: syncx.ForAtomicBool(false),
	//	after:  syncx.ForAtomicBool(false),
	//}
	//	done := syncx.NewDoneChan()
	//	c0 := NewSaramaKafkaConsumer(brokers, topics, groupId, WithConsumerInterceptor(interceptor))
	//	c0.Looper(func(ctx context.Context, message *sarama.ConsumerMessage) error {
	//		t.Logf("consumer0 handle message, key: %s, offset: %d, value: %s", message.Key, message.Offset, message.Value)
	//		time.Sleep(2 * time.Second)
	//		return nil
	//	})
	//
	//	proc.AddShutdownListener(func() {
	//		done.Close()
	//	})
	//
	//	<-done.Done()
}

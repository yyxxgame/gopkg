//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/core/syncx"
)

var (
	brokers = []string{"101.33.209.36:9092"}
	topics  = []string{"testsdk"}
	groupId = "gopkg_test_group"
)

type (
	fakeInterceptor struct {
		before *syncx.AtomicBool
		after  *syncx.AtomicBool
	}
)

func (p *fakeInterceptor) BeforeProduce(_ *sarama.ProducerMessage) {
	fmt.Println("before produce message...")
	p.before.Set(true)
}

func (p *fakeInterceptor) AfterProduce(_ *sarama.ProducerMessage, _ error) {
	fmt.Println("after produce message...")
	p.after.Set(true)
}

func (p *fakeInterceptor) BeforeConsume(_ *sarama.ConsumerMessage) {
	fmt.Println("before consume message...")
	p.before.Set(true)
}

func (p *fakeInterceptor) AfterConsume(_ *sarama.ConsumerMessage, _ error) {
	fmt.Println("after consume message...")
	p.after.Set(true)
}

func TestSaramaKafkaProducer(t *testing.T) {
	interceptor := &fakeInterceptor{
		before: syncx.ForAtomicBool(false),
		after:  syncx.ForAtomicBool(false),
	}
	p := NewSaramaKafkaProducer(brokers, WithProducerInterceptor(interceptor))
	p.Publish(topics[0], "test_message", []byte(stringx.Randn(16)))

	assert.True(t, interceptor.before.True())
	assert.True(t, interceptor.after.True())
}

func TestSaramaKafkaConsumer(t *testing.T) {
	interceptor := &fakeInterceptor{
		before: syncx.ForAtomicBool(false),
		after:  syncx.ForAtomicBool(false),
	}
	done := syncx.NewDoneChan()
	c0 := NewSaramaKafkaConsumer(brokers, topics, groupId, WithConsumerInterceptor(interceptor))
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

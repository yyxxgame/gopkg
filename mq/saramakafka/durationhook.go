//@File     durationhook.go
//@Time     2024/5/14
//@Author   #Suyghur,

package saramakafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/timex"
)

type producerDurationHook struct {
}

type consumerDurationHook struct {
	groupId string
}

func newProducerDurationHook() *producerDurationHook {
	return &producerDurationHook{}
}

func (c *producerDurationHook) Handle(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error {
	topic := message.Topic

	start := timex.Now()

	err := next(ctx, message)

	duration := timex.Since(start)

	metricProduceDur.Observe(duration.Milliseconds(), topic)
	if err != nil {
		metricProduceErr.Inc(topic)
	}

	return err
}

func newConsumerDurationHook(groupId string) *consumerDurationHook {
	return &consumerDurationHook{
		groupId: groupId,
	}

}

func (c *consumerDurationHook) Handle(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error {
	topic := message.Topic

	start := timex.Now()

	err := next(ctx, message)

	duration := timex.Since(start)

	metricConsumeDur.Observe(duration.Milliseconds(), topic, c.groupId)
	if err != nil {
		metricConsumeErr.Inc(topic, c.groupId)
	}

	return err
}

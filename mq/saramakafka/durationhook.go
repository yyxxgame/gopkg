//@File     durationhook.go
//@Time     2024/5/14
//@Author   #Suyghur,

package saramakafka

import (
	"context"

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

func (c *producerDurationHook) Handle(ctx context.Context, topic, key, payload string, next hookFunc) error {
	start := timex.Now()

	err := next(ctx, topic, key, payload)

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

func (c *consumerDurationHook) Handle(ctx context.Context, topic, key, payload string, next hookFunc) error {
	start := timex.Now()

	err := next(ctx, topic, key, payload)

	duration := timex.Since(start)

	metricConsumeDur.Observe(duration.Milliseconds(), topic, c.groupId)
	if err != nil {
		metricConsumeErr.Inc(topic, c.groupId)
	}

	return err
}

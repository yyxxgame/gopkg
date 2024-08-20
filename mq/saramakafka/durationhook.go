//@File     durationhook.go
//@Time     2024/5/14
//@Author   #Suyghur,

package saramakafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/timex"
)

func producerDurationHook() ProducerHook {
	return func(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error {
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
}

func consumerDurationHook(groupId string) ConsumerHook {
	return func(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error {
		topic := message.Topic

		start := timex.Now()

		err := next(ctx, message)

		duration := timex.Since(start)

		metricConsumeDur.Observe(duration.Milliseconds(), topic, groupId)
		if err != nil {
			metricConsumeErr.Inc(topic, groupId)
		}

		return err
	}
}

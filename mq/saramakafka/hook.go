//@File     hook.go
//@Time     2024/5/15
//@Author   #Suyghur,

package saramakafka

import (
	"context"

	"github.com/IBM/sarama"
)

type (
	ProducerHookFunc func(ctx context.Context, message *sarama.ProducerMessage) error

	ConsumerHookFunc func(ctx context.Context, message *sarama.ConsumerMessage) error

	ProducerHook func(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error

	ConsumerHook func(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error
)

func chainProducerHooks(hooks ...ProducerHook) ProducerHook {
	if len(hooks) == 0 {
		return nil
	}
	return func(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error {
		return hooks[0](ctx, message, getProducerHookFunc(hooks, 0, next))

	}
}

func getProducerHookFunc(hooks []ProducerHook, index int, final ProducerHookFunc) ProducerHookFunc {
	if index == len(hooks)-1 {
		return final
	}

	return func(ctx context.Context, message *sarama.ProducerMessage) error {
		return hooks[index+1](ctx, message, getProducerHookFunc(hooks, index+1, final))
	}
}

func chainConsumerHooks(hooks ...ConsumerHook) ConsumerHook {
	if len(hooks) == 0 {
		return nil
	}

	return func(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error {
		return hooks[0](ctx, message, getConsumerHookFunc(hooks, 0, next))
	}
}

func getConsumerHookFunc(hooks []ConsumerHook, index int, final ConsumerHookFunc) ConsumerHookFunc {
	if index == len(hooks)-1 {
		return final
	}

	return func(ctx context.Context, message *sarama.ConsumerMessage) error {
		return hooks[index+1](ctx, message, getConsumerHookFunc(hooks, index+1, final))
	}
}

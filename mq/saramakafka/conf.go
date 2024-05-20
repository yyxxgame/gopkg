//@File     conf.go
//@Time     2024/5/13
//@Author   #Suyghur,

package saramakafka

import (
	"github.com/IBM/sarama"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	OptionConf struct {
		username       string
		password       string
		tracer         oteltrace.Tracer
		partitioner    sarama.PartitionerConstructor
		producerHooks  []ProducerHook
		consumerHooks  []ConsumerHook
		disableStatLag bool
	}

	Option func(c *OptionConf)
)

func WithSaslPlaintext(username, password string) Option {
	return func(c *OptionConf) {
		c.username = username
		c.password = password
	}
}

func WithPartitioner(partitioner sarama.PartitionerConstructor) Option {
	return func(c *OptionConf) {
		c.partitioner = partitioner
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *OptionConf) {
		c.tracer = tracer
	}
}

func WithDisableStatLag() Option {
	return func(c *OptionConf) {
		c.disableStatLag = true
	}
}

func WithProducerHook(hook ProducerHook) Option {
	return func(c *OptionConf) {
		c.producerHooks = append(c.producerHooks, hook)
	}
}

func WithConsumerHook(hook ConsumerHook) Option {
	return func(c *OptionConf) {
		c.consumerHooks = append(c.consumerHooks, hook)
	}
}

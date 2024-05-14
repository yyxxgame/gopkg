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
		username             string
		password             string
		producerInterceptors []ProducerInterceptor
		consumerInterceptors []ConsumerInterceptor
		partitioner          sarama.PartitionerConstructor
		tracer               oteltrace.Tracer
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

func WithProducerInterceptor(interceptor ProducerInterceptor) Option {
	return func(c *OptionConf) {
		c.producerInterceptors = append(c.producerInterceptors, interceptor)
	}
}

func WithConsumerInterceptor(interceptor ConsumerInterceptor) Option {
	return func(c *OptionConf) {
		c.consumerInterceptors = append(c.consumerInterceptors, interceptor)
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *OptionConf) {
		c.tracer = tracer
	}
}

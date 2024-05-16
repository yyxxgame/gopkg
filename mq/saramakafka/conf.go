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
		username      string
		password      string
		tracer        oteltrace.Tracer
		partitioner   sarama.PartitionerConstructor
		enableStatLag bool
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

func WithEnableStatLag() Option {
	return func(c *OptionConf) {
		c.enableStatLag = true
	}
}

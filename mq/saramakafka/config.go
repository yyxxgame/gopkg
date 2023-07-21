//@File     config.go
//@Time     2023/04/29
//@Author   #Suyghur,

package saramakafka

import (
	"github.com/IBM/sarama"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	config struct {
		brokers     []string
		username    string
		password    string
		tracer      oteltrace.Tracer
		partitioner sarama.PartitionerConstructor
	}

	Option func(c *config)
)

func WithSaslPlaintext(username, password string) Option {
	return func(c *config) {
		c.username = username
		c.password = password
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

func WithPartitioner(partitioner sarama.PartitionerConstructor) Option {
	return func(c *config) {
		c.partitioner = partitioner
	}
}

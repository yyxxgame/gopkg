//@File     config.go
//@Time     2023/04/29
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"github.com/IBM/sarama"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	config struct {
		brokers     []string
		username    string
		password    string
		partitioner sarama.PartitionerConstructor
		tracer      oteltrace.Tracer
		metricHook  MetricHook
	}

	MetricHook func(ctx context.Context, topic string)

	Option func(c *config)
)

func WithSaslPlaintext(username, password string) Option {
	return func(c *config) {
		c.username = username
		c.password = password
	}
}

func WithPartitioner(partitioner sarama.PartitionerConstructor) Option {
	return func(c *config) {
		c.partitioner = partitioner
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

func WithMetricHook(metricHook MetricHook) Option {
	return func(c *config) {
		c.metricHook = metricHook
	}
}

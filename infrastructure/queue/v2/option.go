//@File     option.go
//@Time     2024/9/11
//@Author   #Suyghur,

package v2

import (
	"github.com/yyxxgame/gopkg/mq/saramakafka"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	Option func(c *controller)
)

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *controller) {
		c.tracer = tracer
	}
}

func WithHooks(hooks ...saramakafka.ConsumerHook) Option {
	return func(c *controller) {
		c.hooks = append(c.hooks, hooks...)
	}
}

//@File     option.go
//@Time     2024/8/13
//@Author   #Suyghur,

package v2

import (
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

func WithHook(hook Hook) Option {
	return func(c *controller) {
		c.hooks = append(c.hooks, hook)
	}
}

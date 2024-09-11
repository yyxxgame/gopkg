//@File     option.go
//@Time     2024/8/13
//@Author   #Suyghur,

package v2

import (
	"context"

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

func WithHook(hook func(ctx context.Context, next func(ctx context.Context) error) error) Option {
	return func(c *controller) {
		c.hooks = append(c.hooks, hook)
	}
}

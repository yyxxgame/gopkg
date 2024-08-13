//@File     tracer.go
//@Time     2024/8/13
//@Author   #Suyghur,

package internal

import (
	"context"
	"fmt"

	"github.com/yyxxgame/gopkg/xtrace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TracerHook(tracer oteltrace.Tracer, jobName string) func(ctx context.Context, next func(ctx context.Context) error) error {
	return func(ctx context.Context, next func(ctx context.Context) error) error {
		spanName := fmt.Sprintf("%s.%s", "cronTaskController", jobName)

		return xtrace.WithTraceHook(ctx, tracer, oteltrace.SpanKindInternal, spanName, func(ctx context.Context) error {
			return next(ctx)
		})
	}
}

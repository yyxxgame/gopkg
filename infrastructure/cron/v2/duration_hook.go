//@File     duration_hook.go
//@Time     2025/8/19
//@Author   #Suyghur,

package v2

import (
	"context"
	"time"

	"github.com/yyxxgame/gopkg/infrastructure/cron/v2/internal"
	"github.com/zeromicro/go-zero/core/timex"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	DurationHook struct {
		Tracer oteltrace.Tracer
	}
)

func NewDurationHook(tracer oteltrace.Tracer) *DurationHook {
	return &DurationHook{
		Tracer: tracer,
	}
}

func (h *DurationHook) ExecHook() Hook {
	return func(ctx context.Context, job ICronJob, next func(ctx context.Context, job ICronJob) error) error {
		start := timex.Now()

		var err error
		if h.Tracer != nil {
			spanCtx, endSpan := h.startSpan(ctx, job)

			err = next(spanCtx, job)

			endSpan(err)
		} else {
			err = next(ctx, job)
		}

		duration := timex.Since(start)

		internal.MetricExecJobsDur.Observe(duration.Milliseconds(), job.Named())
		if err != nil {
			internal.MetricExecJobsErr.Inc(job.Named())
		}

		internal.MetricExecJobsLastTimestamp.Set(float64(time.Now().Unix()), job.Named())
		return err
	}
}

func (h *DurationHook) startSpan(ctx context.Context, job ICronJob) (context.Context, func(err error)) {
	ctx, span := h.Tracer.Start(ctx, job.Named(), oteltrace.WithSpanKind(oteltrace.SpanKindClient))

	return ctx, func(err error) {
		defer span.End()

		if err == nil {
			span.SetStatus(codes.Ok, "")
			return
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
}

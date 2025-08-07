//@File     duration_hook.go
//@Time     2025/5/28
//@Author   #Suyghur,

package internal

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	v9rds "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/timex"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const spanName = "redis"

var (
	redisCmdsAttributeKey = attribute.Key("redis.cmds")
)

type (
	DurationHook struct {
		Tracer        oteltrace.Tracer
		SlowThreshold *syncx.AtomicDuration
	}
)

func (h *DurationHook) DialHook(next v9rds.DialHook) v9rds.DialHook {
	return next
}

func (h *DurationHook) ProcessHook(next v9rds.ProcessHook) v9rds.ProcessHook {
	return func(ctx context.Context, cmd v9rds.Cmder) error {
		if _, ok := ignoreCmds[cmd.Name()]; ok {
			return next(ctx, cmd)
		}

		start := timex.Now()
		ctx, endSpan := h.startSpan(ctx, cmd)

		err := next(ctx, cmd)

		endSpan(err)
		duration := timex.Since(start)
		if duration > h.SlowThreshold.Load() {
			h.logDuration(ctx, []v9rds.Cmder{cmd}, duration)
			metricSlowCount.Inc(cmd.Name())
		}

		metricReqDur.Observe(duration.Milliseconds(), cmd.Name())
		if msg := h.formatError(err); len(msg) > 0 {
			metricReqErr.Inc(cmd.Name(), msg)
		}
		return err

	}
}

func (h *DurationHook) ProcessPipelineHook(next v9rds.ProcessPipelineHook) v9rds.ProcessPipelineHook {
	return func(ctx context.Context, cmds []v9rds.Cmder) error {
		if len(cmds) <= 0 {
			return next(ctx, cmds)
		}

		for _, item := range cmds {
			if _, ok := ignoreCmds[item.Name()]; ok {
				return next(ctx, cmds)
			}
		}

		start := timex.Now()
		ctx, endSpan := h.startSpan(ctx, cmds...)

		err := next(ctx, cmds)

		endSpan(err)
		duration := timex.Since(start)
		if duration > h.SlowThreshold.Load()*time.Duration(len(cmds)) {
			h.logDuration(ctx, cmds, duration)
		}

		metricReqDur.Observe(duration.Milliseconds(), "Pipeline")
		if msg := h.formatError(err); len(msg) > 0 {
			metricReqErr.Inc("Pipeline", msg)
		}

		return err
	}
}

func (h *DurationHook) startSpan(ctx context.Context, cmds ...v9rds.Cmder) (context.Context, func(err error)) {
	ctx, span := h.Tracer.Start(ctx, spanName, oteltrace.WithSpanKind(oteltrace.SpanKindClient))

	cmdStrs := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		cmdStrs = append(cmdStrs, cmd.Name())
	}
	span.SetAttributes(redisCmdsAttributeKey.StringSlice(cmdStrs))

	return ctx, func(err error) {
		defer span.End()

		if err == nil || errors.Is(err, v9rds.Nil) {
			span.SetStatus(codes.Ok, "")
			return
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
}
func (h *DurationHook) formatError(err error) string {
	if err == nil || errors.Is(err, v9rds.Nil) {
		return ""
	}

	var opErr *net.OpError
	ok := errors.As(err, &opErr)
	if ok && opErr.Timeout() {
		return "timeout"
	}

	switch {
	case errors.Is(err, io.EOF):
		return "eof"
	case errors.Is(err, context.DeadlineExceeded):
		return "context deadline"
	case errors.Is(err, breaker.ErrServiceUnavailable):
		return "breaker open"
	default:
		return "unexpected error"
	}
}

func (h *DurationHook) logDuration(ctx context.Context, cmds []v9rds.Cmder, duration time.Duration) {
	var buf strings.Builder
	for k, cmd := range cmds {
		if k > 0 {
			buf.WriteByte('\n')
		}
		var build strings.Builder
		for i, arg := range cmd.Args() {
			if i > 0 {
				build.WriteByte(' ')
			}
			build.WriteString(mapping.Repr(arg))
		}
		buf.WriteString(build.String())
	}
	logx.WithContext(ctx).WithDuration(duration).Slowf("[REDIS] slowcall on executing: %s", buf.String())
}

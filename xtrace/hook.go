//@File     hook.go
//@Time     2023/04/26
//@Author   #Suyghur,

package xtrace

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type HookFunc func(ctx context.Context) error

func WithTraceHook(ctx context.Context, tracer oteltrace.Tracer, spanKind oteltrace.SpanKind, name string, hook HookFunc, kv ...attribute.KeyValue) error {
	spanCtx, span := tracer.Start(ctx, name, oteltrace.WithSpanKind(spanKind))
	span.SetAttributes(kv...)
	defer span.End()

	err := hook(spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}

func RunWithTraceHook(tracer oteltrace.Tracer, spanKind oteltrace.SpanKind, traceId string, name string, hook HookFunc, kv ...attribute.KeyValue) error {
	var ctx context.Context
	if traceId == "" {
		ctx = context.Background()
	} else {
		traceIdFromHex, _ := oteltrace.TraceIDFromHex(traceId)
		ctx = oteltrace.ContextWithSpanContext(ctx, oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
			TraceID: traceIdFromHex,
		}))
	}

	spanCtx, span := tracer.Start(ctx, name, oteltrace.WithSpanKind(spanKind))
	span.SetAttributes(kv...)
	defer span.End()

	err := hook(spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

//@File     hook.go
//@Time     2023/04/26
//@Author   #Suyghur,

package xtrace

import (
	"context"
	"github.com/yyxxgame/gopkg/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
)

type HookFunc func(ctx context.Context) error

func WithTraceHook(ctx context.Context, tracer oteltrace.Tracer, spanKind oteltrace.SpanKind, name string, hook HookFunc, kv ...attribute.KeyValue) {
	spanCtx, span := tracer.Start(ctx, name, oteltrace.WithSpanKind(spanKind))
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	defer span.End()
	if err := hook(spanCtx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func RunWithTraceHook(tracer oteltrace.Tracer, spanKind oteltrace.SpanKind, traceId string, name string, hook HookFunc, kv ...attribute.KeyValue) {
	propagator := otel.GetTextMapPropagator()
	header := http.Header{}
	if len(traceId) != 0 {
		header.Set("x-trace-id", traceId)
	}
	ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(header))
	spanName := utils.CallerFuncName()
	traceIdFromHex, _ := oteltrace.TraceIDFromHex(traceId)
	ctx = oteltrace.ContextWithSpanContext(ctx, oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID: traceIdFromHex,
	}))
	spanCtx, span := tracer.Start(ctx, spanName, oteltrace.WithSpanKind(spanKind))
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	defer span.End()
	propagator.Inject(spanCtx, propagation.HeaderCarrier(header))
	if err := hook(spanCtx); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, "")

	}
}

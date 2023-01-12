//@File     trace.go
//@Time     2022/05/16
//@Author   #Suyghur,

package xtrace

import (
	"context"
	"github.com/yyxxgame/gopkg/internal/utils"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
)

func StartFuncSpan(ctx context.Context, name string, callback func(context.Context), kv ...attribute.KeyValue) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	spanCtx, span := tracer.Start(ctx, name, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	defer span.End()
	callback(spanCtx)
}

func RunOnTracing(traceId string, name string, callback func(ctx context.Context), kv ...attribute.KeyValue) {
	propagator := otel.GetTextMapPropagator()
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
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
	spanCtx, span := tracer.Start(ctx, spanName, oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	defer span.End()
	propagator.Inject(spanCtx, propagation.HeaderCarrier(header))
	callback(spanCtx)
}

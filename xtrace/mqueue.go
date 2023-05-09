//@File     mqueue.go
//@Time     2023/05/04
//@Author   LvWenQi

package xtrace

import (
	"context"
	"encoding/json"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type HeaderCarrier struct {
	Carrier *propagation.HeaderCarrier `json:"carrier"`
}

func StartMqProducerTrace(ctx context.Context, name string, kv ...attribute.KeyValue) (trace.Span, *HeaderCarrier) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	spanCtx, span := tracer.Start(ctx, name, oteltrace.WithSpanKind(oteltrace.SpanKindProducer))
	carrier := &propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(spanCtx, carrier)
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	return span, &HeaderCarrier{Carrier: carrier}
}

func StartMqConsumerTrace(ctx context.Context, name string, carrierStr string, kv ...attribute.KeyValue) trace.Span {
	var carrier HeaderCarrier
	err := json.Unmarshal([]byte(carrierStr), &carrier)
	if err != nil {
		carrier.Carrier = &propagation.HeaderCarrier{}
	}
	wireCtx := otel.GetTextMapPropagator().Extract(ctx, carrier.Carrier)
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	_, span := tracer.Start(wireCtx, name, oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	return span
}

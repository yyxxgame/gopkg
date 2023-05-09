//@File     mqueue.go
//@Time     2023/05/04
//@Author   LvWenQi

package xtrace

import (
	"context"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type MqMsg struct {
	Carrier *propagation.HeaderCarrier `json:"carrier"`
	Body    string                     `json:"body"`
}

func StartMqProducerTrace(ctx context.Context, name string, kv ...attribute.KeyValue) (trace.Span, *MqMsg) {
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	spanCtx, span := tracer.Start(ctx, name, oteltrace.WithSpanKind(oteltrace.SpanKindProducer))
	carrier := &propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(spanCtx, carrier)
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	return span, &MqMsg{Carrier: carrier}
}

func StartMqConsumerTrace(ctx context.Context, name string, mqMsg *MqMsg, kv ...attribute.KeyValue) trace.Span {
	if mqMsg.Carrier == nil {
		mqMsg.Carrier = &propagation.HeaderCarrier{}
	}
	wireCtx := otel.GetTextMapPropagator().Extract(ctx, mqMsg.Carrier)
	tracer := otel.GetTracerProvider().Tracer(gozerotrace.TraceName)
	_, span := tracer.Start(wireCtx, name, oteltrace.WithSpanKind(oteltrace.SpanKindConsumer))
	span.AddEvent(name, oteltrace.WithAttributes(kv...))
	return span
}

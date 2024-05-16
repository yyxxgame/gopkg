//@File     tracehook.go
//@Time     2024/5/14
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"fmt"

	"github.com/yyxxgame/gopkg/xtrace"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const spanName = "saramakafka"

var (
	attributeTopic   = attribute.Key("saramakafka.topic")
	attributeKey     = attribute.Key("saramakafka.key")
	attributePayload = attribute.Key("saramakafka.payload")
)

type traceHook struct {
	tracer   oteltrace.Tracer
	spanKind oteltrace.SpanKind
}

func newTraceHook(tracer oteltrace.Tracer, spanKind oteltrace.SpanKind) *traceHook {
	return &traceHook{
		tracer:   tracer,
		spanKind: spanKind,
	}
}

func (c *traceHook) Handle(ctx context.Context, topic, key, payload string, next hookFunc) error {
	name := fmt.Sprintf("%s.%s", spanName, c.spanKind.String())
	return xtrace.WithTraceHook(ctx, c.tracer, c.spanKind, name, func(ctx context.Context) error {
		return next(ctx, topic, key, payload)
	}, attributeTopic.String(topic), attributeKey.String(key), attributePayload.String(payload))
}

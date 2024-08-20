//@File     tracehook.go
//@Time     2024/5/14
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
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

func producerTraceHook(tracer oteltrace.Tracer) ProducerHook {
	return func(ctx context.Context, message *sarama.ProducerMessage, next ProducerHookFunc) error {
		bKey, _ := message.Key.Encode()
		key := string(bKey)

		bPayload, _ := message.Key.Encode()
		payload := string(bPayload)

		name := fmt.Sprintf("%s.%s", spanName, oteltrace.SpanKindProducer.String())

		return xtrace.WithTraceHook(ctx, tracer, oteltrace.SpanKindProducer, name, func(ctx context.Context) error {
			return next(ctx, message)
		}, attributeTopic.String(message.Topic), attributeKey.String(key), attributePayload.String(payload))
	}
}

func consumerTraceHook(tracer oteltrace.Tracer) ConsumerHook {
	return func(ctx context.Context, message *sarama.ConsumerMessage, next ConsumerHookFunc) error {
		key := string(message.Key)

		payload := string(message.Value)

		name := fmt.Sprintf("%s.%s", spanName, oteltrace.SpanKindConsumer.String())

		return xtrace.WithTraceHook(ctx, tracer, oteltrace.SpanKindConsumer, name, func(ctx context.Context) error {
			return next(ctx, message)
		}, attributeTopic.String(message.Topic), attributeKey.String(key), attributePayload.String(payload))
	}
}

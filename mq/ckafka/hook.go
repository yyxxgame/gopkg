//@File     hook.go
//@Time     2023/04/24
//@Author   #Suyghur,

package ckafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel/attribute"
)

var (
	spanName           = "ckafka"
	ckafkaAttributeKey = attribute.Key("ckafka.message")
)

type IHook interface {
	startSpan(ctx context.Context, message *kafka.Message) context.Context
	endSpan(ctx context.Context, err error)
}

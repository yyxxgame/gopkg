//@File     ext.go
//@Time     2023/04/28
//@Author   #Suyghur,

package ckafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yyxxgame/gopkg/mq"
)

func GetTraceIdFromHeader(message *kafka.Message) string {
	traceId := ""
	for _, header := range message.Headers {
		if header.Key == mq.TraceId {
			traceId = string(header.Value)
		}
	}
	return traceId
}

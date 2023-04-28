//@File     ext.go
//@Time     2023/04/28
//@Author   #Suyghur,

package ckafka

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

func GetTraceIdFromHeader(message *kafka.Message) string {
	traceId := ""
	for _, header := range message.Headers {
		if header.Key == ckafkaTraceIdKey {
			traceId = string(header.Value)
		}
	}
	return traceId
}

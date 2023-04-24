//@File     consumer.go
//@Time     2023/04/24
//@Author   #Suyghur,

package ckafka

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

// see: https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
var defaultConsumerConfigMap = &kafka.ConfigMap{
	"acks": 1,
	// 请求发生错误时重试次数，建议将该值设置为大于0，失败重试最大程度保证消息不丢失
	"retries": 0,
	//"api.version.request": "false",
	//"broker.version.fallback": "0.10.0",
	// 发送请求失败时到下一次重试请求之间的时间
	"retry.backoff.ms": 1000,
	// producer 网络请求的超时时间。
	"socket.timeout.ms": 5000,
	// 设置客户端内部重试间隔。
	"reconnect.backoff.max.ms": 2000,
	// 消息最大载荷10m
	"message.max.bytes": 10485760,
	"security.protocol": "plaintext",
}

type (
	entry struct {
		idx int32
		*kafka.Consumer
		deliveryChs []chan kafka.Event
	}

	consumer struct {
		entries []*entry
	}
)

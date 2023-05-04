//@File     config.go
//@Time     2023/04/25
//@Author   #Suyghur,

package ckafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const ckafkaTraceKey = "ckafka-key"
const ckafkaTracePayload = "ckafka-payload"

// see: https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
var (
	defaultProducerConfigMap = &kafka.ConfigMap{
		"acks": 1,
		// 请求发生错误时重试次数，建议将该值设置为大于0，失败重试最大程度保证消息不丢失
		"retries": 3,
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
	defaultConsumerConfigMap = &kafka.ConfigMap{
		"auto.offset.reset":        "latest",
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"max.poll.interval.ms":     300000,
		"heartbeat.interval.ms":    3000,
		"session.timeout.ms":       6000,
		"auto.commit.interval.ms":  5000,
		"broker.address.family":    "v4",
		"security.protocol":        "plaintext",
	}
)

type (
	instance struct {
		Username  string
		Password  string
		configMap *kafka.ConfigMap
		tracer    oteltrace.Tracer
		*baseProducer
	}
	baseProducer struct {
		// 生产者分区策略
		partitioner IPartitioner
	}

	Option func(i *instance)
)

func WithSaslPlaintext(username, password string) Option {
	return func(i *instance) {
		i.Username = username
		i.Password = password
	}
}

func WithConfigMap(configMap *kafka.ConfigMap) Option {
	return func(i *instance) {
		i.configMap = configMap
	}
}

func WithProducerPartitioner(partitioner IPartitioner) Option {
	return func(i *instance) {
		i.partitioner = partitioner
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(i *instance) {
		i.tracer = tracer
	}
}

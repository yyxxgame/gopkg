//@File     producer.go
//@Time     2023/04/21
//@Author   #Suyghur,

package ckafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"strings"
)

// see: https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
var defaultProducerConfigMap = &kafka.ConfigMap{
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
	IProducer interface {
		registerCallback()
		Emit(key string, bMsg []byte)
		EmitCtx(ctx context.Context, key string, bMsg []byte)
		Release()
	}

	// A Producer wraps a kafka.Producer and traces its operations.
	producer struct {
		*kafka.Producer

		Username string
		Password string

		configMap *kafka.ConfigMap

		topic string

		numPartition  int32
		nextPartition int32
		partitioner   IPartitioner

		deliveryChan chan kafka.Event
	}
	Option func(p *producer)
)

// NewCkafkaProducer calls kafka.NewProducer and wraps the resulting Producer with
// tracing instrumentation.
func NewCkafkaProducer(brokers []string, topic string, opts ...Option) IProducer {
	impl := &producer{}
	for _, opt := range opts {
		opt(impl)
	}
	if impl.configMap == nil {
		impl.configMap = defaultProducerConfigMap
	}
	_ = impl.configMap.SetKey("bootstrap.servers", strings.Join(brokers, ","))
	if impl.Username != "" && impl.Password != "" {
		_ = impl.configMap.SetKey("security.protocol", "SASL_PLAINTEXT")
		_ = impl.configMap.SetKey("sasl.mechanisms", "PLAIN")
		_ = impl.configMap.SetKey("sasl.username", impl.Username)
		_ = impl.configMap.SetKey("sasl.password", impl.Password)
	}
	if p, err := kafka.NewProducer(impl.configMap); err != nil {
		logx.Errorf("ckafka.NewProducer on error: %v", err)
		panic(err)
	} else {
		impl.Producer = p
	}

	adminCli, err := kafka.NewAdminClientFromProducer(impl.Producer)
	if err != nil {
		logx.Errorf("ckafka.NewAdminClientFromProducer on error: %v", err)
		panic(err)
	}

	metadata, err := adminCli.GetMetadata(&topic, false, 10000)
	if err != nil {
		logx.Errorf("ckafka.GetMetadata on error: %v", err)
		panic(err)
	}

	if topicMetadata, ok := metadata.Topics[topic]; !ok {
		err = fmt.Errorf("topic#%s no topic metadata", topic)
		logx.Errorf("ckafka.Metadata.Topics on error: %v", err)
		panic(err)
	} else {
		if len(topicMetadata.Partitions) == 0 {
			err = fmt.Errorf("topic#%s no topic partitions", topic)
			logx.Errorf("ckafka.Metadata.Topics on error: %v", err)
			panic(err)
		} else {
			impl.numPartition = int32(len(topicMetadata.Partitions))
		}
	}

	impl.topic = topic
	impl.deliveryChan = make(chan kafka.Event)

	if impl.partitioner == nil {
		impl.partitioner = NewRandomPartitioner(topic)
	}

	impl.nextPartition = kafka.PartitionAny

	impl.registerCallback()
	return impl
}

func WithSaslPlaintext(username, passwod string) Option {
	return func(p *producer) {
		p.Username = username
		p.Password = passwod
	}
}

func WithConfigMap(configMap *kafka.ConfigMap) Option {
	return func(p *producer) {
		p.configMap = configMap
	}
}

func WithPartitioner(partitioner IPartitioner) Option {
	return func(p *producer) {
		p.partitioner = partitioner
	}
}

func (p *producer) registerCallback() {
	gopool.Go(func() {
		for event := range p.Events() {
			switch message := event.(type) {
			case *kafka.Message:
				if message.TopicPartition.Error == nil {
					logx.Infof("ckafka produce success: %s, message: %s", message.TopicPartition.String(), string(message.Value))
				} else {
					logx.Errorf("ckafka produce fail: %s, message: %s", message.TopicPartition.String(), string(message.Value))
				}
			case kafka.Error:
				logx.Errorf("ckafka notice error: %s", message.Error())
			default:
				logx.Infof("ckafka ignored event: %s", message)
			}
		}
	})
}
func (p *producer) Emit(key string, bMsg []byte) {
	p.EmitCtx(context.Background(), key, bMsg)
}

func (p *producer) EmitCtx(ctx context.Context, key string, bMsg []byte) {
	traceId := xtrace.GetTraceId(ctx).String()
	message := &kafka.Message{}
	message.TopicPartition = kafka.TopicPartition{
		Topic:     &p.topic,
		Partition: p.nextPartition,
	}
	message.Key = []byte(key)
	message.Value = bMsg
	message.Headers = []kafka.Header{{
		Key:   "ckafka-trace-id",
		Value: []byte(traceId),
	}}
	ctx = p.startSpan(ctx, message)
	err := p.Produce(message, p.deliveryChan)

	e := <-p.deliveryChan
	ev := e.(*kafka.Message)
	if ev.TopicPartition.Error != nil {
		err = ev.TopicPartition.Error
	}
	p.nextPartition, err = p.partitioner.Partition(message, p.numPartition)
	defer p.endSpan(ctx, err)
}

func (p *producer) Release() {
	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
	p.Close()
}

func (p *producer) startSpan(ctx context.Context, message *kafka.Message) context.Context {
	tracer := trace.TracerFromContext(ctx)
	ctx, span := tracer.Start(ctx, spanName, oteltrace.WithSpanKind(oteltrace.SpanKindProducer))
	span.SetAttributes(ckafkaAttributeKey.String(message.String()))
	return ctx
}

func (p *producer) endSpan(ctx context.Context, err error) {
	span := oteltrace.SpanFromContext(ctx)
	defer span.End()

	if err == nil {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}

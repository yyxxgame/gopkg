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
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"strings"
)

type (
	IProducer interface {
		install()
		Emit(key string, bMsg []byte)
		EmitCtx(ctx context.Context, key string, bMsg []byte)
		produceMessage(message *kafka.Message) error
		Release()
	}

	// A Producer wraps a kafka.Producer and traces its operations.
	producer struct {
		*instance
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
	//ProducerOption func(p *producer)
)

// NewCkafkaProducer calls kafka.NewProducer and wraps the resulting Producer with
// tracing instrumentation.
func NewCkafkaProducer(brokers []string, topic string, opts ...Option) IProducer {
	impl := &producer{
		instance: &instance{},
	}

	for _, opt := range opts {
		opt(impl.instance)
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

	impl.install()
	return impl
}

func (p *producer) install() {
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
		Key:   ckafkaTraceIdKey,
		Value: []byte(traceId),
	}}
	if p.tracer != nil {
		xtrace.WithTraceHook(ctx, p.tracer, oteltrace.SpanKindProducer, "ckafka.EmitCtx", func(ctx context.Context) error {
			return p.produceMessage(message)
		},
			attribute.String(ckafkaTraceKey, key),
			attribute.String(ckafkaTracePayload, message.String()),
		)
	} else {
		p.produceMessage(message)
	}
}

func (p *producer) produceMessage(message *kafka.Message) error {
	if err := p.Produce(message, p.deliveryChan); err != nil {
		logx.Errorf("ckafka.EmitCtx.Produce on error: %v", err)
		return err
	}
	e := <-p.deliveryChan
	ev := e.(*kafka.Message)
	if err := ev.TopicPartition.Error; err != nil {
		logx.Errorf("ckafka.EmitCtx.TopicPartition on error: %v", err)
		return err
	}
	p.nextPartition, _ = p.partitioner.Partition(message, p.numPartition)
	return nil
}

func (p *producer) Release() {
	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
	p.Close()
}

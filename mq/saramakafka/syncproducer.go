//@File     syncproducer.go
//@Time     2023/04/29
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/yyxxgame/gopkg/mq"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	IProducer interface {
		Publish(topic, key string, bMsg []byte) error
		PublishCtx(ctx context.Context, topic, key string, bMsg []byte) error
		Release()
	}

	syncProducer struct {
		*config
		sarama.SyncProducer
	}
)

func NewSaramaSyncProducer(brokers []string, opts ...Option) IProducer {
	p := &syncProducer{
		config: &config{},
	}
	for _, opt := range opts {
		opt(p.config)
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	if p.partitioner == nil {
		p.partitioner = sarama.NewRoundRobinPartitioner
	}
	config.Producer.Partitioner = p.partitioner

	if p.username != "" && p.password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = p.username
		config.Net.SASL.Password = p.password
	} else {
		config.Net.SASL.Enable = false
	}

	config.Version = sarama.V1_0_0_0

	p.brokers = brokers

	// Initialize the client
	if client, err := sarama.NewClient(p.brokers, config); err != nil {
		logx.Errorf("saramakafka.NewSaramaSyncProducer.NewClient error: %v", err)
		panic(err)
	} else {
		if producer, err := sarama.NewSyncProducerFromClient(client); err != nil {
			logx.Errorf("saramakafka.NewSaramaSyncProducer.NewSyncProducerFromClient error: %v", err)
			panic(err)
		} else {
			p.SyncProducer = producer
		}
	}
	return p
}

func (p *syncProducer) Publish(topic, key string, bMsg []byte) error {
	return p.PublishCtx(context.Background(), topic, key, bMsg)
}

func (p *syncProducer) PublishCtx(ctx context.Context, topic, key string, bMsg []byte) error {
	traceId := xtrace.GetTraceId(ctx).String()
	message := &sarama.ProducerMessage{}
	message.Key = sarama.StringEncoder(key)
	message.Topic = topic
	message.Value = sarama.StringEncoder(bMsg)

	if p.tracer != nil && traceId != "" {
		traceHeader := sarama.RecordHeader{
			Key:   sarama.ByteEncoder(mq.TraceId),
			Value: sarama.ByteEncoder(traceId),
		}
		message.Headers = []sarama.RecordHeader{traceHeader}
		return xtrace.WithTraceHook(ctx, p.tracer, oteltrace.SpanKindProducer, "saramakafka.PublishCtx.SendMessage", func(ctx context.Context) error {
			return p.publishMessage(message)
		},
			attribute.String(mq.TraceMqTopic, topic),
			attribute.String(mq.TraceMqKey, key),
			attribute.String(mq.TraceMqPayload, string(bMsg)))
	} else {
		return p.publishMessage(message)
	}
}

func (p *syncProducer) publishMessage(message *sarama.ProducerMessage) error {
	if partition, offset, err := p.SendMessage(message); err != nil {
		logx.Errorf("saramakafka.PublishCtx.SendMessage to topic: %s, on error: %v", message.Topic, err)
		return err
	} else {
		logx.Infof("saramakafka.PublishCtx.SendMessage to topic: %s, on success, partition: %d, offset: %v", message.Topic, partition, offset)
		return nil
	}
}

func (p *syncProducer) Release() {
	_ = p.Close()
}

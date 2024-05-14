//@File     producer.go
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

	producer struct {
		*OptionConf
		sarama.SyncProducer
	}
)

func NewSaramaKafkaProducer(brokers []string, opts ...Option) IProducer {
	p := &producer{
		OptionConf: &OptionConf{
			producerInterceptors: []ProducerInterceptor{},
		},
	}
	for _, opt := range opts {
		opt(p.OptionConf)
	}

	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	if p.partitioner == nil {
		p.partitioner = sarama.NewRoundRobinPartitioner
	}
	config.Producer.Partitioner = p.partitioner
	config.Producer.Interceptors = append(config.Producer.Interceptors, p)

	if p.username == "" || p.password == "" {
		config.Net.SASL.Enable = false
	} else {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = p.username
		config.Net.SASL.Password = p.password
	}

	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logx.Errorf("[SARAMA-KAFKA-ERROR]: MustNewProducer on error: %v", err)
		panic(err)
	}

	p.SyncProducer = syncProducer

	return p
}

func (p *producer) Publish(topic, key string, bMsg []byte) error {
	return p.PublishCtx(context.Background(), topic, key, bMsg)
}

func (p *producer) PublishCtx(ctx context.Context, topic, key string, bMsg []byte) error {
	traceId := xtrace.GetTraceId(ctx).String()
	message := &sarama.ProducerMessage{}
	message.Key = sarama.StringEncoder(key)
	message.Topic = topic
	message.Value = sarama.ByteEncoder(bMsg)

	if p.tracer == nil || traceId == "" {
		return p.publishMessage(ctx, message)
	}

	traceHeader := sarama.RecordHeader{
		Key:   sarama.ByteEncoder(mq.HeaderKey),
		Value: sarama.ByteEncoder(traceId),
	}
	message.Headers = []sarama.RecordHeader{traceHeader}
	return xtrace.WithTraceHook(ctx, p.tracer, oteltrace.SpanKindProducer, "saramakafka.publishMessage", func(ctx context.Context) error {
		return p.publishMessage(ctx, message)
	},
		attribute.String(mq.HeaderTopic, topic),
		attribute.String(mq.HeaderKey, key),
		attribute.String(mq.HeaderPayload, string(bMsg)))
}

func (p *producer) publishMessage(ctx context.Context, message *sarama.ProducerMessage) error {
	partition, offset, err := p.SendMessage(message)

	if err != nil {
		logx.WithContext(ctx).Errorf("[SARAMA-KAFKA-ERROR]: publishMessage.SendMessage to topic: %s, on error: %v", message.Topic, err)
		return err
	}

	logx.WithContext(ctx).Infof("[SARAMA-KAFKA]: publishMessage.SendMessage to topic: %s, on success, partition: %d, offset: %v", message.Topic, partition, offset)

	return nil
}

func (p *producer) Release() {
	_ = p.Close()
}

func (p *producer) OnSend(message *sarama.ProducerMessage) {
	for _, interceptor := range p.producerInterceptors {
		interceptor(message)
	}
}

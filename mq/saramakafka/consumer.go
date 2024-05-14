//@File     consumer.go
//@Time     2023/04/30
//@Author   #Suyghur,

package saramakafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/yyxxgame/gopkg/mq"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	IConsumer interface {
		Looper(handler ConsumerHandler)
		Release()
	}

	consumer struct {
		*OptionConf
		topics []string
		sarama.ConsumerGroup
		handler ConsumerHandler
	}

	ConsumerHandler func(ctx context.Context, message *sarama.ConsumerMessage) error
)

func NewSaramaKafkaConsumer(brokers, topics []string, groupId string, opts ...Option) IConsumer {
	c := &consumer{
		OptionConf: &OptionConf{
			consumerInterceptors: []ConsumerInterceptor{},
		},
		topics: topics,
	}

	for _, opt := range opts {
		opt(c.OptionConf)
	}

	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Retry.Max = 99
	config.Consumer.Offsets.AutoCommit.Enable = true

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		logx.Errorf("[SARAMA-KAFKA-ERROR]: NewSaramaKafkaConsumer on error: %v", err)
		panic(err)
	}

	c.ConsumerGroup = consumerGroup

	return c
}

func (c *consumer) Looper(handler ConsumerHandler) {
	c.handler = handler
	gopool.Go(func() {
		for {
			if err := c.Consume(context.Background(), c.topics, c); err != nil {
				logx.Error(err.Error())
				panic(err.Error())
			}
		}
	})
}
func (c *consumer) Release() {
	c.PauseAll()
	_ = c.Close()
}

func (c *consumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			for _, topic := range c.topics {

				if topic != message.Topic {
					continue
				}

				traceId := c.getHeaderValue(mq.HeaderTraceId, message.Headers)
				if c.tracer == nil || traceId == "" {
					_ = c.handleMessage(context.Background(), session, message)
				} else {
					_ = xtrace.RunWithTraceHook(c.tracer, oteltrace.SpanKindConsumer, traceId, "saramakafka.ConsumeClaim", func(ctx context.Context) error {
						return c.handleMessage(ctx, session, message)
					},
						attribute.String(mq.HeaderTopic, message.Topic),
						attribute.String(mq.HeaderKey, string(message.Key)),
						attribute.String(mq.HeaderPayload, string(message.Value)))
				}
			}
		case <-session.Context().Done():
			// Should return when `session.Context()` is done.
			// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
			// https://github.com/IBM/sarama/issues/1192
			return nil
		}
	}
}

func (c *consumer) handleMessage(ctx context.Context, session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) error {
	c.beforeProduceMessage(message)
	err := c.handler(ctx, message)
	c.afterProduceMessage(message, err)

	if err != nil {
		logx.WithContext(ctx).Errorf("[SARAMA-KAFKA-ERROR]: ConsumeClaim.handleMessage on error: %v", err)
		return err
	}

	session.MarkMessage(message, "")
	return nil
}

func (c *consumer) beforeProduceMessage(message *sarama.ConsumerMessage) {
	for _, interceptor := range c.consumerInterceptors {
		interceptor.BeforeConsume(message)
	}
}

func (c *consumer) afterProduceMessage(message *sarama.ConsumerMessage, err error) {
	for _, interceptor := range c.consumerInterceptors {
		interceptor.AfterConsume(message, err)
	}
}

func (c *consumer) getHeaderValue(header mq.Header, headers []*sarama.RecordHeader) string {
	for _, item := range headers {
		if string(item.Key) == header.String() {
			return string(item.Value)
		}
	}
	return ""
}

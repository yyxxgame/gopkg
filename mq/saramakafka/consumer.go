//@File     consumer.go
//@Time     2023/04/30
//@Author   #Suyghur,

package saramakafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/yyxxgame/gopkg/mq"
	"github.com/yyxxgame/gopkg/mq/saramakafka/internal"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/zeromicro/go-zero/core/logx"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	IConsumer interface {
		Looper(handler ConsumerHandler)
		Release()
	}

	consumer struct {
		*OptionConf
		topics    []string
		hooks     []internal.Hook
		finalHook internal.Hook
		sarama.ConsumerGroup
		handler ConsumerHandler
	}

	ConsumerHandler func(ctx context.Context, message *sarama.ConsumerMessage) error
)

func NewSaramaKafkaConsumer(brokers, topics []string, groupId string, opts ...Option) IConsumer {
	c := &consumer{
		OptionConf: &OptionConf{},
		topics:     topics,
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

	if c.tracer != nil {
		c.hooks = append(c.hooks, internal.NewTraceHook(c.tracer, oteltrace.SpanKindConsumer).Handle)
	}

	c.hooks = append(c.hooks, internal.NewDurationHook(oteltrace.SpanKindConsumer).Handle)

	c.finalHook = internal.ChainHooks(c.hooks...)

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
		case message := <-claim.Messages():
			for _, topic := range c.topics {
				if topic != message.Topic {
					continue
				}

				var ctx context.Context
				traceId := c.getHeaderValue(mq.HeaderTraceId, message.Headers)
				if traceId == "" {
					ctx = context.Background()
				} else {
					traceIdFromHex, _ := oteltrace.TraceIDFromHex(traceId)
					ctx = oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
						TraceID: traceIdFromHex,
					}))
				}

				_ = c.finalHook(ctx, message.Topic, string(message.Key), string(message.Value), func(ctx context.Context, topic, key, payload string) error {
					return c.handleMessage(ctx, session, message)
				})
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
	err := c.handler(ctx, message)
	if err != nil {
		logx.WithContext(ctx).Errorf("[SARAMA-KAFKA-ERROR]: ConsumeClaim.handleMessage on error: %v", err)
		return err
	}
	session.MarkMessage(message, "")
	return nil
}

func (c *consumer) getHeaderValue(header mq.Header, headers []*sarama.RecordHeader) string {
	for _, item := range headers {
		if string(item.Key) == header.String() {
			return string(item.Value)
		}
	}
	return ""
}

//@File     consumer.go
//@Time     2023/04/30
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/yyxxgame/gopkg/mq"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"sync"
)

type (
	IConsumer interface {
		Looper(handler ConsumerHandler)
		LooperSync(handler ConsumerHandler)
		Release()
	}

	consumer struct {
		sarama.ConsumerGroup
		groupId string
		topics  []string
		handler ConsumerHandler
		wg      sync.WaitGroup
		runSync bool
		*config
	}

	ConsumerHandler func(ctx context.Context, message *sarama.ConsumerMessage) error
)

func NewSaramaConsumer(brokers, topics []string, groupId string, opts ...Option) IConsumer {
	c := &consumer{
		config: &config{},
	}

	for _, opt := range opts {
		opt(c.config)
	}

	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Retry.Max = 99
	config.Consumer.Offsets.AutoCommit.Enable = true

	if cg, err := sarama.NewConsumerGroup(brokers, groupId, config); err != nil {
		logx.Errorf("saramakafka.NewConsumerGroupAuto.NewConsumerGroup error: %v", err)
		panic(err)
	} else {
		c.ConsumerGroup = cg
	}
	c.topics = topics
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

func (c *consumer) LooperSync(handler ConsumerHandler) {
	c.wg.Add(1)
	c.Looper(handler)
	c.runSync = true
	c.wg.Wait()
}

func (c *consumer) Release() {
	if c.runSync {
		c.wg.Done()
	}
	c.PauseAll()
	c.Close()
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		for _, topic := range c.topics {
			if topic == msg.Topic {
				traceId := GetTraceIdFromHeader(msg.Headers)
				if c.tracer != nil && traceId != "" {
					return xtrace.RunWithTraceHook(c.tracer, oteltrace.SpanKindConsumer, traceId, "saramakafka.ConsumeClaim", func(ctx context.Context) error {
						return c.handleMessageCtx(ctx, session, msg)
					},
						attribute.String(mq.TraceMqTopic, msg.Topic),
						attribute.String(mq.TraceMqKey, string(msg.Value)),
						attribute.String(mq.TraceMqPayload, string(msg.Value)))
				} else {
					return c.handleMessage(session, msg)
				}
			}
		}
	}
	return nil
}

func (c *consumer) handleMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	return c.handleMessageCtx(context.Background(), session, msg)
}

func (c *consumer) handleMessageCtx(ctx context.Context, session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	if err := c.handler(ctx, msg); err != nil {
		logx.WithContext(ctx).Errorf("saramakafka.ConsumeClaim.handleMessage on error: %v", err)
		return err
	} else {
		session.MarkMessage(msg, "")
	}
	return nil
}

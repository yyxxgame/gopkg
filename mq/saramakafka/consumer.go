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
	"sync"
)

type (
	IConsumer interface {
		Looper(handler ConsumerHandler)
		LooperSync(handler ConsumerHandler)
		Release()
	}

	consumer struct {
		client  sarama.Client
		group   sarama.ConsumerGroup
		groupId string
		topics  []string
		handler ConsumerHandler
		wg      sync.WaitGroup
		runSync bool
		*config
		enableBroadcastModel bool
		resetOffsetOnce      sync.Once
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

	if client, err := sarama.NewClient(brokers, config); err != nil {
		logx.Errorf("saramakafka.NewSaramaConsumer.NewClient error: %v", err)
		panic(err)
	} else {
		if cg, err := sarama.NewConsumerGroupFromClient(groupId, client); err != nil {
			logx.Errorf("saramakafka.NewSaramaConsumer.NewConsumerGroupFromClient error: %v", err)
			panic(err)
		} else {
			c.client = client
			c.group = cg
		}
	}

	c.topics = topics
	return c
}

func NewSaramaBroadcastConsumer(brokers, topics []string, groupId string, opts ...Option) IConsumer {
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

	if client, err := sarama.NewClient(brokers, config); err != nil {
		logx.Errorf("saramakafka.NewSaramaBroadcastConsumer.NewClient error: %v", err)
		panic(err)
	} else {
		if cg, err := sarama.NewConsumerGroupFromClient(groupId, client); err != nil {
			logx.Errorf("saramakafka.NewSaramaBroadcastConsumer.NewConsumerGroupFromClient error: %v", err)
			panic(err)
		} else {
			c.client = client
			c.group = cg
		}
	}

	c.topics = topics
	c.enableBroadcastModel = true
	return c
}

func (c *consumer) Looper(handler ConsumerHandler) {
	c.handler = handler
	gopool.Go(func() {
		for {
			if err := c.group.Consume(context.Background(), c.topics, c); err != nil {
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
	c.group.PauseAll()
	_ = c.group.Close()
	_ = c.client.Close()
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	if !c.enableBroadcastModel {
		return nil
	}
	c.resetOffsetOnce.Do(func() {
		for topic, partitions := range session.Claims() {
			for _, partition := range partitions {
				if offset, err := c.client.GetOffset(topic, partition, sarama.OffsetNewest); err != nil {
					continue
				} else {
					logx.Infof("reset offset on setup: %d", offset)
					session.MarkOffset(topic, partition, offset, "")
				}
			}
		}
	})
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

				traceId := GetTraceIdFromHeader(message.Headers)
				if c.tracer == nil || traceId == "" {
					_ = c.handleMessage(session, message)
				} else {
					_ = xtrace.RunWithTraceHook(c.tracer, oteltrace.SpanKindConsumer, traceId, "saramakafka.ConsumeClaim.handleMessage", func(ctx context.Context) error {
						return c.handleMessage(session, message)
					},
						attribute.String(mq.TraceMqTopic, message.Topic),
						attribute.String(mq.TraceMqKey, string(message.Key)),
						attribute.String(mq.TraceMqPayload, string(message.Value)))
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

func (c *consumer) handleMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	return c.handleMessageCtx(context.Background(), session, msg)
}

func (c *consumer) handleMessageCtx(ctx context.Context, session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {
	err := c.handler(ctx, msg)

	if err != nil {
		logx.WithContext(ctx).Errorf("saramakafka.ConsumeClaim.handleMessage on error: %v", err)
		return err
	}

	session.MarkMessage(msg, "")

	defer func() {
		if c.metricHook != nil {
			c.metricHook(ctx, msg.Topic)
		}
	}()

	return nil
}

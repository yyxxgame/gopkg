//@File     consumer.go
//@Time     2023/04/30
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/yyxxgame/gopkg/mq"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/zeromicro/go-zero/core/fx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/timex"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	IConsumer interface {
		Looper(handler ConsumerHandler)
		Release()
	}

	consumer struct {
		*OptionConf
		done       *syncx.DoneChan
		hasDone    *syncx.AtomicBool
		groupId    string
		topics     []string
		hooks      []hook
		finalHook  hook
		statTicker timex.Ticker
		sarama.Client
		sarama.ClusterAdmin
		sarama.ConsumerGroup
		handler ConsumerHandler
	}

	ConsumerHandler func(ctx context.Context, message *sarama.ConsumerMessage) error
)

func NewSaramaKafkaConsumer(brokers, topics []string, groupId string, opts ...Option) IConsumer {
	c := &consumer{
		OptionConf: &OptionConf{},
		done:       syncx.NewDoneChan(),
		hasDone:    syncx.ForAtomicBool(false),
		groupId:    groupId,
		topics:     topics,
		statTicker: timex.NewTicker(time.Second * 60),
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

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		logx.Errorf("[SARAMA-KAFKA-ERROR]: NewSaramaKafkaConsumer on error: %v", err)
		panic(err)
	}
	c.Client = client

	clusterAdmin, err := sarama.NewClusterAdminFromClient(c.Client)
	if err != nil {
		logx.Errorf("[SARAMA-KAFKA-ERROR]: NewSaramaKafkaConsumer on error: %v", err)
		panic(err)
	}
	c.ClusterAdmin = clusterAdmin

	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupId, c.Client)
	if err != nil {
		logx.Errorf("[SARAMA-KAFKA-ERROR]: NewSaramaKafkaConsumer on error: %v", err)
		panic(err)
	}

	c.ConsumerGroup = consumerGroup

	if c.tracer != nil {
		c.hooks = append(c.hooks, newTraceHook(c.tracer, oteltrace.SpanKindConsumer).Handle)
	}

	c.hooks = append(c.hooks, newConsumerDurationHook(groupId).Handle)

	c.finalHook = chainHooks(c.hooks...)

	gopool.Go(func() {
		for {
			select {
			case <-c.statTicker.Chan():
				c.statLag()
			case <-c.done.Done():
				c.statTicker.Stop()
				c.hasDone.Set(true)
				return
			}
		}
	})

	return c
}

func (c *consumer) Looper(handler ConsumerHandler) {
	c.handler = handler
	gopool.Go(func() {
		for {
			err := c.Consume(context.Background(), c.topics, c)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				logx.Errorf("[SARAMA-KAFKA-ERROR]: Consume on error: %v", err)
				panic(err.Error())
			}

			if c.hasDone.True() {
				return
			}
		}
	})
}
func (c *consumer) Release() {
	c.done.Close()
	c.ConsumerGroup.PauseAll()
	_ = c.ConsumerGroup.Close()
	_ = c.ClusterAdmin.Close()
	_ = c.Client.Close()
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

				ctx := context.Background()
				traceId := c.getHeaderValue(mq.HeaderTraceId, message.Headers)
				if traceId != "" {
					traceIdFromHex, _ := oteltrace.TraceIDFromHex(traceId)
					ctx = oteltrace.ContextWithSpanContext(ctx, oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
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

func (c *consumer) statLag() {
	if c.disableStatLag {
		return
	}

	fx.From(func(source chan<- any) {
		for _, item := range c.topics {
			source <- item
		}
	}).Parallel(func(item any) {
		topic := item.(string)
		var total int64
		partitions, err := c.Client.Partitions(topic)
		if err != nil {
			return
		}

		consumerGroupOffsets, err := c.ClusterAdmin.ListConsumerGroupOffsets(c.groupId, map[string][]int32{topic: partitions})
		if err != nil {
			return
		}
		for _, partition := range partitions {
			latestOffset, _ := c.GetOffset(topic, partition, sarama.OffsetNewest)
			offset := consumerGroupOffsets.Blocks[topic][partition].Offset
			lag := latestOffset - offset
			total += lag
			metricLag.Set(float64(lag), topic, c.groupId, strconv.FormatInt(int64(partition), 10))
		}
		metricLagSum.Set(float64(total), topic, c.groupId)
	})
}

func (c *consumer) getHeaderValue(header mq.Header, headers []*sarama.RecordHeader) string {
	for _, item := range headers {
		if string(item.Key) == header.String() {
			return string(item.Value)
		}
	}
	return ""
}

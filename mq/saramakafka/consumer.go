//@File     consumer.go
//@Time     2023/04/30
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"errors"
	"slices"
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
		Loop()
		Release()
	}

	consumer struct {
		*OptionConf
		done       *syncx.DoneChan
		hasDone    *syncx.AtomicBool
		groupId    string
		topics     []string
		hooks      []ConsumerHook
		finalHook  ConsumerHook
		statTicker timex.Ticker
		sarama.Client
		sarama.ClusterAdmin
		sarama.ConsumerGroup
		handler ConsumerHandler
	}

	ConsumerHandler func(ctx context.Context, message *sarama.ConsumerMessage) error
)

func NewConsumer(brokers, topics []string, groupId string, handler ConsumerHandler, opts ...Option) IConsumer {
	c := &consumer{
		OptionConf: &OptionConf{
			consumerHooks: []ConsumerHook{},
		},
		done:       syncx.NewDoneChan(),
		hasDone:    syncx.ForAtomicBool(false),
		groupId:    groupId,
		topics:     topics,
		hooks:      []ConsumerHook{},
		statTicker: timex.NewTicker(time.Second * 30),
		handler:    handler,
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

	if c.username == "" || c.password == "" {
		config.Net.SASL.Enable = false
	} else {
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.Version = sarama.SASLHandshakeV0
		config.Net.SASL.Handshake = true
		config.Net.SASL.User = c.username
		config.Net.SASL.Password = c.password

		config.Net.TLS.Enable = false
	}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		logx.Errorf("[SARAMA-KAFKA-ERROR]: NewConsumer on error: %v", err)
		panic(err)
	}
	c.Client = client

	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupId, c.Client)
	if err != nil {
		panic(err)
	}
	c.ConsumerGroup = consumerGroup

	clusterAdmin, err := sarama.NewClusterAdminFromClient(c.Client)
	if err != nil {
		panic(err)
	}
	c.ClusterAdmin = clusterAdmin

	if c.tracer != nil {
		c.hooks = append(c.hooks, consumerTraceHook(c.tracer))
	}

	c.hooks = append(c.hooks, consumerDurationHook(groupId))

	c.hooks = append(c.hooks, c.consumerHooks...)

	c.finalHook = chainConsumerHooks(c.hooks...)

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

func (c *consumer) Loop() {
	gopool.Go(func() {
		for {
			if c.hasDone.True() {
				return
			}

			err := c.Consume(context.Background(), c.topics, c)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				panic(err.Error())
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
			if !slices.Contains(c.topics, message.Topic) {
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

			_ = c.finalHook(ctx, message, func(ctx context.Context, message *sarama.ConsumerMessage) error {
				err := c.handler(ctx, message)
				if err != nil {
					return err
				}
				session.MarkMessage(message, "")
				return nil
			})
		case <-session.Context().Done():
			// Should return when `session.Context()` is done.
			// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
			// https://github.com/IBM/sarama/issues/1192
			return nil
		}
	}
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

//@File     consumer.go
//@Time     2023/04/24
//@Author   #Suyghur,

package ckafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/lang"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/threading"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"strings"
)

type (
	IConsumer interface {
		Looper(handler ConsumerHandler)
		Release()
	}

	consumer struct {
		*instance
		*kafka.Consumer
		Username  string
		Password  string
		configMap *kafka.ConfigMap
		signal    chan lang.PlaceholderType
		done      *syncx.AtomicBool
	}

	ConsumerHandler func(message *kafka.Message) error
)

func NewCkafkaConsumer(brokers, topics []string, groupId string, opts ...Option) IConsumer {
	impl := &consumer{
		instance: &instance{},
	}

	for _, opt := range opts {
		opt(impl.instance)
	}
	if impl.configMap == nil {
		impl.configMap = defaultConsumerConfigMap
	}
	_ = impl.configMap.SetKey("bootstrap.servers", strings.Join(brokers, ","))
	_ = impl.configMap.SetKey("group.id", groupId)
	if impl.Username != "" && impl.Password != "" {
		_ = impl.configMap.SetKey("security.protocol", "SASL_PLAINTEXT")
		_ = impl.configMap.SetKey("sasl.mechanisms", "PLAIN")
		_ = impl.configMap.SetKey("sasl.username", impl.Username)
		_ = impl.configMap.SetKey("sasl.password", impl.Password)
	}

	if c, err := kafka.NewConsumer(impl.configMap); err != nil {
		logx.Errorf("ckafka.NewConsumer on error: %v", err)
		panic(err)
	} else {
		impl.Consumer = c
		impl.signal = make(chan lang.PlaceholderType)
		impl.done = syncx.NewAtomicBool()
		if err := impl.SubscribeTopics(topics, nil); err != nil {
			logx.Errorf("ckafka.SubscribeTopics on error : %v", err)
			panic(err)
		}
	}
	return impl
}

func (c *consumer) Looper(handler ConsumerHandler) {
	threading.GoSafe(func() {
		for !c.done.True() {
			select {
			case <-c.signal:
				c.done.Set(true)
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}
				switch e := ev.(type) {
				case *kafka.Message:
					traceId := GetTraceIdFromHeader(e)
					if c.tracer != nil && traceId != "" {
						_ = xtrace.RunWithTraceHook(c.tracer, oteltrace.SpanKindConsumer, traceId, "ckafka.Looper.handleMessage", func(ctx context.Context) error {
							return c.handleMessage(e, handler)
						},
							attribute.String(ckafkaTraceKey, string(e.Key)),
							attribute.String(ckafkaTracePayload, e.String()),
						)
					} else {
						_ = c.handleMessage(e, handler)
					}
				case kafka.AssignedPartitions:
					_ = c.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					_ = c.Unassign()
				case kafka.Error:
					logx.Errorf("ckafka consumer notice error, code: %d, msg: %s", e.Code(), e.Error())
				default:
					logx.Infof("ckafka consumer ignored event: %v", e.String())
				}
			}
		}
	})
}

func (c *consumer) handleMessage(message *kafka.Message, handler ConsumerHandler) error {
	if err := handler(message); err != nil {
		logx.Errorf("ckafka.Looper.onMessage on error: %v", err)
		return err
	} else {
		if _, err = c.StoreMessage(message); err != nil {
			logx.Errorf("ckafka.Looper.StoreMessage on error: %v", err)
			return err
		} else {
			if _, err := c.CommitMessage(message); err != nil {
				logx.Errorf("ckafka.Looper.StoreMessage on error: %v", err)
				return err
			}
		}
	}
	return nil
}

func (c *consumer) Release() {
	c.signal <- lang.Placeholder
	_ = c.Close()
}

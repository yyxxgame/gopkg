//@File     consumer.go
//@Time     2023/07/17
//@Author   #Suyghur,

package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
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
		Looper(topic string, handler ConsumerHandler)
		LooperSync(topic string, handler ConsumerHandler)
		handleMessage(ctx context.Context, task *asynq.Task) error
		Release()
	}

	consumer struct {
		*asynq.Server
		*asynq.ServeMux
		*config
		Topic   string
		handler ConsumerHandler
		wg      sync.WaitGroup
		runSync bool
	}

	ConsumerHandler func(ctx context.Context, message *Message) error
)

func NewRmqConsumer(broker, topic string, opts ...Option) IConsumer {
	c := &consumer{
		config: &config{},
		Topic:  topic,
	}
	for _, opt := range opts {
		opt(c.config)
	}

	clientOpt := asynq.RedisClientOpt{
		Addr: broker,
	}
	if c.password != "" {
		clientOpt.Password = c.password
	}

	clientConfig := asynq.Config{
		// Optionally specify multiple queues with different priority.
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	}
	if c.workerNum > 10 {
		clientConfig.Concurrency = c.workerNum
	} else {
		clientConfig.Concurrency = 10
	}

	c.Server = asynq.NewServer(clientOpt, clientConfig)
	c.ServeMux = asynq.NewServeMux()
	return c
}

func (c *consumer) Looper(topic string, handler ConsumerHandler) {
	c.handler = handler
	formatTopic := fmt.Sprintf("%s:%s", "rmq", topic)
	c.ServeMux.HandleFunc(formatTopic, c.handleMessage)
	gopool.Go(func() {
		if err := c.Server.Run(c.ServeMux); err != nil {
			logx.Errorf("rmq server run on error: %v", err.Error())
			panic(err.Error())
		}
	})
}

func (c *consumer) LooperSync(topic string, handler ConsumerHandler) {
	c.wg.Add(1)
	c.Looper(topic, handler)
	c.runSync = true
	c.wg.Wait()
}

func (c *consumer) handleMessage(ctx context.Context, task *asynq.Task) error {
	message := &Message{}
	_ = json.Unmarshal(task.Payload(), message)
	traceId := GetTraceIdFromHeader(message.Headers)
	if c.tracer != nil && traceId != "" {
		return xtrace.RunWithTraceHook(c.tracer, oteltrace.SpanKindConsumer, traceId, "rmq.handleMessage", func(ctx context.Context) error {
			return c.handler(ctx, message)
		},
			attribute.String(mq.TraceMqTopic, string(message.Topic)),
			attribute.String(mq.TraceMqPayload, string(message.Payload)))
	} else {
		return c.handler(ctx, message)
	}
}

func (c *consumer) Release() {
	if c.runSync {
		c.wg.Done()
	}
	c.Server.Shutdown()
}

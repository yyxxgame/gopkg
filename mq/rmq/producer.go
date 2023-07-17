//@File     producer.go
//@Time     2023/07/17
//@Author   #Suyghur,

package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/yyxxgame/gopkg/mq"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	IProducer interface {
		Publish(topic string, bMsg []byte) error
		PublishCtx(ctx context.Context, topic string, bMsg []byte) error
		Release()
	}

	producer struct {
		*asynq.Client
		*asynq.RedisClientOpt
		*config
	}
)

func NewRmqProducer(broker string, opts ...Option) IProducer {
	p := &producer{
		config: &config{},
	}
	for _, opt := range opts {
		opt(p.config)
	}
	clientOpt := asynq.RedisClientOpt{
		Addr: broker,
	}
	if p.password != "" {
		clientOpt.Password = p.password
	}

	p.Client = asynq.NewClient(clientOpt)
	return p
}

func (p *producer) Publish(topic string, bMsg []byte) error {
	return p.PublishCtx(context.Background(), topic, bMsg)
}

func (p *producer) PublishCtx(ctx context.Context, topic string, bMsg []byte) error {
	formatTopic := fmt.Sprintf("%s:%s", "rmq", topic)
	traceId := xtrace.GetTraceId(ctx).String()
	message := &Message{
		Topic:   []byte(topic),
		Payload: bMsg,
	}

	if p.tracer != nil && traceId != "" {
		header := &Header{
			Key:   []byte(mq.TraceId),
			Value: []byte(traceId),
		}
		message.Headers = append(message.Headers, header)
		bPayload, _ := json.Marshal(message)
		task := asynq.NewTask(formatTopic, bPayload)
		return xtrace.WithTraceHook(ctx, p.tracer, oteltrace.SpanKindProducer, "rmq.PublishCtx.SendMessage", func(ctx context.Context) error {
			if _, err := p.EnqueueContext(ctx, task); err != nil {
				logx.WithContext(ctx).Errorf("rmq PublishCtx on error: %v", err)
				return err
			} else {
				return nil
			}
		},
			attribute.String(mq.TraceMqTopic, topic),
			attribute.String(mq.TraceMqPayload, string(bMsg)))
	} else {
		bPayload, _ := json.Marshal(message)
		task := asynq.NewTask(formatTopic, bPayload)
		if _, err := p.EnqueueContext(ctx, task); err != nil {
			logx.WithContext(ctx).Errorf("rmq PublishCtx on error: %v", err)
			return err
		} else {
			return nil
		}
	}
}

func (p *producer) Release() {
	_ = p.Close()
}

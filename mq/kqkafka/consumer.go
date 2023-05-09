//@File     consumer.go
//@Time     2023/05/08
//@Author   LvWenQi

package kqkafka

import (
	"context"
	"fmt"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	RunHandle func(ctx context.Context, key, value string) error

	Consumer struct {
		conf    kq.KqConf
		handler kq.ConsumeHandle
	}

	ConsumerInst interface {
		Start()
		Stop()
	}
)

func NewConsumer(conf kq.KqConf, handler RunHandle) ConsumerInst {
	return &Consumer{
		conf: conf,
		handler: func(k, v string) error {
			ctx := context.Background()
			span := xtrace.StartMqConsumerTrace(
				ctx, fmt.Sprintf("%s.%s", conf.Group, conf.Topic), k, attribute.String("params", v),
			)
			defer span.End()
			return handler(trace.ContextWithSpan(ctx, span), k, v)

		},
	}
}

func (sel *Consumer) Start() {
	threading.GoSafe(func() {
		q := kq.MustNewQueue(sel.conf, kq.WithHandle(sel.handler))
		defer q.Stop()
		q.Start()
	})
	logx.Infof("[kqkafka.Consumer.Start] conf:%#v", sel.conf)
}

func (sel *Consumer) Stop() {
	logx.Info("[kqkafka.Consumer.Stop] stop, conf:", sel.conf)
}

//@File     controller_context.go
//@Time     2024/9/11
//@Author   #Suyghur,

package v2

import (
	"context"

	"github.com/yyxxgame/gopkg/mq/saramakafka"
	"github.com/zeromicro/go-zero/core/logx"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	IQueueTaskController interface {
		Start()
		Stop()
		RegisterJobs(jobs ...IQueueJob)
	}

	IQueueJob interface {
		OnConsume(ctx context.Context, payload string, params map[string]any) error
		Named() string
	}

	controller struct {
		conf   QueueTaskConf
		tracer oteltrace.Tracer
		jobs   map[string]*WrapperJob
		hooks  []saramakafka.ConsumerHook
	}
)

func NewQueueTaskController(conf QueueTaskConf, opts ...Option) IQueueTaskController {
	c := &controller{
		conf:  conf,
		jobs:  make(map[string]*WrapperJob),
		hooks: []saramakafka.ConsumerHook{},
	}

	for _, opt := range opts {
		opt(c)
	}
	//

	//c.hooks = append(c.hooks, internal.DurationHook)

	return c
}

func (c *controller) Start() {
}

func (c *controller) Stop() {
}

func (c *controller) RegisterJobs(jobs ...IQueueJob) {
	for _, job := range jobs {
		for _, item := range c.conf.Jobs {
			if !item.Enable {
				continue
			}

			if item.Name != job.Named() {
				continue
			}

			if _, exists := c.jobs[item.Name]; exists {
				logx.Errorf("[QUEUE-TASK-CONTROLLER-ERROR]: register job: %s on duplicated error, skip it ...", item.Name)
				continue
			}

			wrapper := &WrapperJob{
				queueJob: job,
				params:   item.Params,
				topic:    item.Topic,
			}

			if c.tracer == nil {
				wrapper.consumer = saramakafka.NewConsumer(
					c.conf.Brokers,
					[]string{item.Topic},
					item.Name,
					wrapper.Consume,
					saramakafka.WithSaslPlaintext(c.conf.Username, c.conf.Password),
					saramakafka.WithConsumerHook(c.hooks...),
				)
			} else {
				wrapper.consumer = saramakafka.NewConsumer(
					c.conf.Brokers,
					[]string{item.Topic},
					item.Name,
					wrapper.Consume,
					saramakafka.WithSaslPlaintext(c.conf.Username, c.conf.Password),
					saramakafka.WithTracer(c.tracer),
					saramakafka.WithConsumerHook(c.hooks...),
				)
			}
			wrapper.Loop()

			c.jobs[item.Name] = wrapper

			logx.Infof("[QUEUE-TASK-CONTROLLER]: register job: %s on success", item.Name)
		}
	}
}

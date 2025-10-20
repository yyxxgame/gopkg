//@File     controller_context.go
//@Time     2024/9/11
//@Author   #Suyghur,

package v2

import (
	"context"
	"fmt"

	"github.com/yyxxgame/gopkg/mq/saramakafka"
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
		jobs   map[string][]*WrapperJob
		hooks  []saramakafka.ConsumerHook
	}
)

func NewQueueTaskController(conf QueueTaskConf, opts ...Option) IQueueTaskController {
	c := &controller{
		conf:  conf,
		jobs:  make(map[string][]*WrapperJob),
		hooks: []saramakafka.ConsumerHook{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *controller) Start() {
	for _, job := range c.jobs {
		for _, item := range job {
			item.Loop()
		}
	}
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
				continue
			}

			for i := 1; i <= item.WorkerNum; i++ {
				wrapperJob := c.prepareJob(job, item)
				c.jobs[item.Name] = append(c.jobs[item.Name], wrapperJob)
			}
			fmt.Printf("[QUEUE-TASK-CONTROLLER]: register job: %s on success\n", item.Name)
		}
	}
}

func (c *controller) prepareJob(job IQueueJob, conf QueueJobConf) *WrapperJob {
	wrapper := &WrapperJob{
		queueJob: job,
		params:   conf.Params,
		topic:    conf.Topic,
	}

	if c.tracer == nil {
		wrapper.consumer = saramakafka.NewConsumer(
			c.conf.Brokers,
			[]string{conf.Topic},
			conf.Name,
			wrapper.Consume,
			saramakafka.WithSaslPlaintext(c.conf.Username, c.conf.Password),
			saramakafka.WithConsumerHook(c.hooks...),
		)
	} else {
		wrapper.consumer = saramakafka.NewConsumer(
			c.conf.Brokers,
			[]string{conf.Topic},
			conf.Name,
			wrapper.Consume,
			saramakafka.WithSaslPlaintext(c.conf.Username, c.conf.Password),
			saramakafka.WithTracer(c.tracer),
			saramakafka.WithConsumerHook(c.hooks...),
		)
	}

	return wrapper
}

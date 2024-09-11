//@File     job.go
//@Time     2024/9/11
//@Author   #Suyghur,

package v2

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/yyxxgame/gopkg/mq/saramakafka"
)

type (
	WrapperJob struct {
		queueJob IQueueJob
		params   map[string]any
		topic    string
		consumer saramakafka.IConsumer
	}
)

func (job *WrapperJob) Consume(ctx context.Context, message *sarama.ConsumerMessage) error {
	// TODO: 这里似乎没有必要判断 topic
	if message.Topic != job.topic {
		return nil
	}

	payload := string(message.Value)
	return job.queueJob.OnConsume(ctx, payload, job.params)
}

func (job *WrapperJob) Loop() {
	job.consumer.Loop()
}

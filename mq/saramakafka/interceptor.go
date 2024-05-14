//@File     interceptor.go
//@Time     2024/5/14
//@Author   #Suyghur,

package saramakafka

import "github.com/IBM/sarama"

type (
	ProducerInterceptor interface {
		BeforeProduce(message *sarama.ProducerMessage)
		AfterProduce(message *sarama.ProducerMessage, err error)
	}

	ConsumerInterceptor interface {
		BeforeConsume(message *sarama.ConsumerMessage)
		AfterConsume(message *sarama.ConsumerMessage, err error)
	}
)

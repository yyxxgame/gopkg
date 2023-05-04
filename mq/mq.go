//@File     mq.go
//@Time     2023/04/29
//@Author   #Suyghur,

package mq

const (
	TraceId           = "x-trace-id"
	TraceKafkaTopic   = "kafka-topic"
	TraceKafkaKey     = "kafka-key"
	TraceKafkaPayload = "kafka-payload"
)

type (
//IProducer interface {
//	Publish(topic, key string, bMsg []byte) error
//	PublishCtx(ctx context.Context, topic, key string, bMsg []byte) error
//	Release()
//}

//	IConsumer interface {
//		Looper[T CkafkaConsumerHandler | SaramaKafkaConsumerHandler](handler T)
//		Release()
//	}
//
// CkafkaConsumerHandler func(message *kafka.Message) error
//
// SaramaKafkaConsumerHandler func(message *sarama.ConsumerMessage) error
//
// producer struct {
// }
//
// consumer struct {
// }
// ProducerOption func(p *producer)
//
// ConsumerOption func(p *producer)
)

//func ConsumerHandler[T *kafka.Message | *sarama.ConsumerMessage](message T) error {
//	return nil
//}
//
//func MustNewProducer() IProducer {
//	return nil
//}
//
//func MustNewConsumer() IConsumer {
//	return nil
//}

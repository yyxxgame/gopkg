//@File     saramakafka_test.go
//@Time     2023/05/04
//@Author   #Suyghur,

package saramakafka

import (
	"context"
	"github.com/IBM/sarama"
	"testing"
	"time"
)

var (
	brokers = []string{"localhost:9092"}
	topics  = []string{"test_sarama_kafka"}
	groupId = "test_group"
)

func TestSaramaKafkaProducer(t *testing.T) {
	p := NewSaramaSyncProducer(brokers)
	p.Publish(topics[0], "test_message", []byte("test_value111"))
}

func TestSaramaKafkaConsumer(t *testing.T) {
	c := NewSaramaConsumer(brokers, topics, groupId)
	c.LooperSync(func(ctx context.Context, message *sarama.ConsumerMessage) error {
		t.Logf("handle message, key: %s, value: %s", message.Key, message.Value)
		time.Sleep(2 * time.Second)
		return nil
	})

	defer c.Release()
}

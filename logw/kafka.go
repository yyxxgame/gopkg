//@File     kafka.go
//@Time     2024/8/20
//@Author   #Suyghur,

package logw

import (
	"strings"

	"github.com/yyxxgame/gopkg/mq/saramakafka"
)

type (
	KafkaWriter struct {
		producer saramakafka.IProducer
		topic    string
	}
)

func MustNewKafkaWriter(producer saramakafka.IProducer, topic string) *KafkaWriter {
	return &KafkaWriter{
		producer: producer,
		topic:    topic,
	}
}

func (w *KafkaWriter) Write(p []byte) (n int, err error) {
	// writing log with newlines, trim them.
	if err := w.producer.Publish(w.topic, "", strings.TrimSpace(string(p))); err != nil {
		return 0, err
	}

	return len(p), nil
}

//@File     kafka_test.go
//@Time     2024/8/20
//@Author   #Suyghur,

package logw

import (
	"errors"
	"testing"

	"github.com/yyxxgame/gopkg/mq/saramakafka"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "test_kafka_writer"
)

func TestKafkaWriter(t *testing.T) {
	p := saramakafka.NewProducer(brokers)
	w := logx.NewWriter(MustNewKafkaWriter(p, topic))

	logx.SetWriter(w)

	logx.Infof("test info message")
	logx.Slowf("test slow message")
	logx.Errorf("test error message with err: %v", errors.New("some cause"))
}

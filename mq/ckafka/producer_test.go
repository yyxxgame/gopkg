//@File     producer_test.go
//@Time     2023/04/23
//@Author   #Suyghur,

package ckafka

import (
	"github.com/zeromicro/go-zero/core/stringx"
	"strconv"
	"testing"
	"time"
)

func TestNewCkafkaProducer(t *testing.T) {
	brokers := []string{"localhost:9092"}
	p := NewCkafkaProducer(brokers, "test_ckafka")
	for i := 0; i < 10; i++ {
		key := "key_" + strconv.Itoa(i)
		msg := []byte(stringx.Randn(20))
		p.Emit(key, msg)
		time.Sleep(time.Second)
	}
}

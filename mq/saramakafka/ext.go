//@File     extdata.go
//@Time     2022/05/17
//@Author   #Suyghur,

package saramakafka

import (
	"github.com/Shopify/sarama"
	"github.com/yyxxgame/gopkg/mq"
)

func GetTraceIdFromHeader(headers []*sarama.RecordHeader) string {
	for _, h := range headers {
		if string(h.Key) == mq.TraceId {
			return string(h.Value)
		}
	}
	return ""
}

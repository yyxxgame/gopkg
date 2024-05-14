//@File     extdata.go
//@Time     2022/05/17
//@Author   #Suyghur,

package saramakafka

import (
	"github.com/IBM/sarama"
	"github.com/yyxxgame/gopkg/mq"
)

func GetHeaderValue(header mq.Header, headers []*sarama.RecordHeader) string {
	for _, item := range headers {
		if string(item.Key) == header.String() {
			return string(item.Value)
		}
	}
	return ""
}

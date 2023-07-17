//@File     message.go
//@Time     2023/07/17
//@Author   #Suyghur,

package rmq

import (
	"github.com/yyxxgame/gopkg/mq"
)

type (
	Message struct {
		Headers []*Header `json:"header"`
		Topic   []byte    `json:"topic"`
		Payload []byte    `json:"payload"`
	}
	Header struct {
		Key   []byte `json:"key"`
		Value []byte `json:"value"`
	}
)

func GetTraceIdFromHeader(headers []*Header) string {
	for _, h := range headers {
		if string(h.Key) == mq.TraceId {
			return string(h.Value)
		}
	}
	return ""
}

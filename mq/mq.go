//@File     mq.go
//@Time     2023/04/29
//@Author   #Suyghur,

package mq

type Header string

const (
	HeaderTraceId = "x-mq-trace-id"
	HeaderTopic   = "x-mq-topic"
	HeaderKey     = "x-mq-key"
	HeaderPayload = "x-mq-payload"
)

func (h Header) String() string {
	return string(h)
}

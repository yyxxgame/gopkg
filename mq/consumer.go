//@File     consumer.go
//@Time     2023/01/03
//@Author   #Suyghur,

package mq

type (
	IConsumer interface {
		Register()
		Close()
	}
)

//@File     producer.go
//@Time     2023/01/03
//@Author   #Suyghur,

package mq

import "context"

type IProducer interface {
	Close()
	Produce(ctx context.Context, key string, message string) error
}

type (
	Option   func(impl *producer)
	producer struct {
	}
)

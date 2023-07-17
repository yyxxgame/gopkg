//@File     config.go
//@Time     2023/07/17
//@Author   #Suyghur,

package rmq

import (
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	config struct {
		password  string
		workerNum int
		tracer    oteltrace.Tracer
	}

	Option func(c *config)
)

func WithPassword(password string) Option {
	return func(c *config) {
		c.password = password
	}
}

func WithWorkerNum(workerNum int) Option {
	return func(c *config) {
		c.workerNum = workerNum
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

//@File     : options.go
//@Time     : 2024/3/6
//@Auther   : Kaishin

package queue

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
)

type (
	OptionService func(ctx *Service)
)

func WithTaskRoute(taskRoute map[string]string) OptionService {
	return func(ctx *Service) {
		ctx.taskRoute = taskRoute
	}
}

func WithKafkaConf(kafkaConf []kq.KqConf) OptionService {
	return func(ctx *Service) {
		ctx.kafkaConf = kafkaConf
	}
}

func WithTelemetry(telemetry trace.Config) OptionService {
	return func(ctx *Service) {
		ctx.telemetry = telemetry
	}
}

func WithObject(obj interface{}) OptionService {
	return func(ctx *Service) {
		ctx.obj = obj
	}
}

func WithFuncRegister(cb FuncRegister) OptionService {
	return func(ctx *Service) {
		ctx.funcRegister = cb
	}
}

//@File     : options.go
//@Time     : 2024/3/6
//@Auther   : Kaishin

package core

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
)

type (
	OptionTaskServerCtx func(ctx *TaskServerCtx)
)

func WithTaskRoute(taskRoute map[string]string) OptionTaskServerCtx {
	return func(ctx *TaskServerCtx) {
		ctx.taskRoute = taskRoute
	}
}

func WithKafkaConf(kafkaConf []kq.KqConf) OptionTaskServerCtx {
	return func(ctx *TaskServerCtx) {
		ctx.kafkaConf = kafkaConf
	}
}

func WithTelemetry(telemetry trace.Config) OptionTaskServerCtx {
	return func(ctx *TaskServerCtx) {
		ctx.telemetry = telemetry
	}
}

func WithObject(obj interface{}) OptionTaskServerCtx {
	return func(ctx *TaskServerCtx) {
		ctx.obj = obj
	}
}

func WithFuncRegister(cb FuncRegister) OptionTaskServerCtx {
	return func(ctx *TaskServerCtx) {
		ctx.funcRegister = cb
	}
}

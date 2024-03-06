//@Author   : KaiShin
//@Time     : 2021/11/18

package core

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
)

type (
	BaseTask struct {
		Topic string
		ITask
	}

	TaskFactory struct {
		ITaskFactory
		Slot  map[string]FuncCreateTask
		Tasks map[string]ITask
	}

	FuncCreateTask func(topic string, serverCtx ITaskServerCtx) ITask

	FuncRegister func(factory ITaskFactory)

	TaskServerCtx struct {
		ITaskServerCtx
		taskRoute    map[string]string
		kafkaConf    []kq.KqConf
		telemetry    trace.Config
		obj          interface{}
		funcRegister FuncRegister
	}
)

//@Author   : KaiShin
//@Time     : 2021/11/18

package core

import (
	"context"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
)

type (
	InstTask interface {
		Run(ctx context.Context, k, v string) error
		Stop()
	}

	BaseTask struct {
		Topic string
		InstTask
	}

	TaskFactory struct {
		ITaskFactory
		Slot  map[string]FuncCreateTask
		Tasks map[string]InstTask
	}

	FuncCreateTask func(topic string, serverCtx ITaskServerCtx) InstTask

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

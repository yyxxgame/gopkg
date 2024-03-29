//@File     : facade.go
//@Time     : 2024/3/6
//@Auther   : Kaishin

package core

import (
	"context"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
)

type (
	ITaskServerCtx interface {
		TaskRoute() map[string]string
		KafkaConf() []kq.KqConf
		Telemetry() trace.Config
		FuncRegister() FuncRegister
		Object() interface{}
	}

	ITaskFactory interface {
		Gen(key string, t FuncCreateTask)
	}

	ITask interface {
		Run(ctx context.Context, k, v string) error
		Stop()
	}
)

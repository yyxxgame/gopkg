//@Author   : KaiShin
//@Time     : 2021/11/18

package queue

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
)

type (
	BaseTask struct {
		Topic string
		ITask
	}

	FuncTask func(topic string, svr IService) ITask

	FuncRegister func(svr IService)

	Service struct {
		IService
		taskRoute    map[string]string
		kafkaConf    []kq.KqConf
		telemetry    trace.Config
		obj          interface{}
		funcRegister FuncRegister
		taskFunc     map[string]FuncTask
		tasks        map[string]ITask
	}
)

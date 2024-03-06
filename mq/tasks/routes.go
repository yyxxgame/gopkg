//@Author   : KaiShin
//@Time     : 2021/11/16

package core

import (
	"fmt"
	"github.com/yyxxgame/gopkg/exception"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/trace"
	"sync"

	"github.com/yyxxgame/gopkg/mq/kqkafka"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var (
	onceFactory sync.Once
	instFactory *TaskFactory
)

func getFactory() ITaskFactory {
	onceFactory.Do(func() {
		instFactory = new(TaskFactory)
		instFactory.Slot = map[string]FuncCreateTask{}
		instFactory.Tasks = map[string]ITask{}
	})

	return instFactory
}

func (sel *TaskFactory) Gen(key string, t FuncCreateTask) {
	sel.Slot[key] = t
	logx.Infof("[TaskFactory.Gen] task:%s", key)
}

func RegisterTasks(serverCtx ITaskServerCtx, group *service.ServiceGroup) {
	funcRegister := serverCtx.FuncRegister()
	if funcRegister == nil {
		logx.Errorf("[task.RegisterTasks] funcRegister is nil, check [gen.go]")
		return
	}
	funcRegister(getFactory())

	mapRegFunc := serverCtx.TaskRoute()
	for _, kafkaConf := range serverCtx.KafkaConf() {
		exception.Try(func() {
			routeKey := fmt.Sprintf("%s.%s", kafkaConf.Topic, kafkaConf.Group)
			funcName, ok := mapRegFunc[routeKey]
			if !ok {
				// kafka topic没有指定消费的task类型
				logx.Errorf(fmt.Sprintf(
					"[task.RegisterTasks] kafka[%s] of consumer[%s] has no task to route, check yaml[TaskRoute]",
					kafkaConf.Topic, routeKey))
				return
			}
			taskTemp, ok := instFactory.Slot[funcName]
			if !ok {
				// 指定的task类型没有注册
				logx.Errorf(fmt.Sprintf("[task.RegisterTasks] func{%s} not register, check [gen.go]", funcName))
				return
			}
			task := taskTemp(kafkaConf.Topic, serverCtx)
			instFactory.Tasks[funcName] = task
			kafkaConf.Telemetry = serverCtx.Telemetry()
			consumer := kqkafka.NewConsumer(kafkaConf, task.Run)
			group.Add(consumer)
			logx.Infof("[task.RegisterTasks] reg func:%s, routeKey:%s", funcName, routeKey)
		}).Catch(func(e exception.Exception) {
			logx.Errorf("[task.RegisterTasks][ERROR] err:%s", e)
		}).Finally(func() {

		})
	}
}

func StopTasks() {
	if instFactory == nil {
		return
	}
	for _, v := range instFactory.Tasks {
		v.Stop()
	}
	logx.Info("Stop Tasks")
}

func NewTaskServerCtx(options ...OptionTaskServerCtx) ITaskServerCtx {
	obj := new(TaskServerCtx)
	for _, option := range options {
		option(obj)
	}
	return obj
}

func (sel *TaskServerCtx) TaskRoute() map[string]string {
	return sel.taskRoute
}

func (sel *TaskServerCtx) KafkaConf() []kq.KqConf {
	return sel.kafkaConf
}

func (sel *TaskServerCtx) Telemetry() trace.Config {
	return sel.telemetry
}

func (sel *TaskServerCtx) Object() interface{} {
	return sel.obj
}

func (sel *TaskServerCtx) FuncRegister() FuncRegister {
	return sel.funcRegister
}

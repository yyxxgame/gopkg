//@Author   : KaiShin
//@Time     : 2021/11/16

package queue

import (
	"fmt"
	"github.com/yyxxgame/gopkg/exception"
	"github.com/yyxxgame/gopkg/mq/kqkafka"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

func NewService(group *service.ServiceGroup, options ...OptionService) IService {
	svc := createSvc()
	for _, option := range options {
		option(svc)
	}
	svc.register(group)
	return svc
}

func (sel *Service) Gen(key string, t FuncTask) {
	sel.taskFunc[key] = t
	logx.Infof("[queue.Service.Gen] task:%s", key)
}

func (sel *Service) Object() interface{} {
	return sel.obj
}

func (sel *Service) Stop() {
	for _, v := range sel.tasks {
		v.Stop()
	}
	logx.Info("Stop tasks")
}

func createSvc() *Service {
	svc := new(Service)
	svc.taskFunc = map[string]FuncTask{}
	svc.tasks = map[string]ITask{}
	return svc
}

func (sel *Service) register(group *service.ServiceGroup) {
	funcRegister := sel.funcRegister
	if funcRegister == nil {
		logx.Errorf("[queue.Service.register] funcRegister is nil, check [gen.go]")
		return
	}
	funcRegister(sel)

	mapRegFunc := sel.taskRoute
	for _, kafkaConf := range sel.kafkaConf {
		exception.Try(func() {
			routeKey := fmt.Sprintf("%s.%s", kafkaConf.Topic, kafkaConf.Group)
			funcName, ok := mapRegFunc[routeKey]
			if !ok {
				// kafka topic没有指定消费的task类型
				logx.Errorf(fmt.Sprintf(
					"[queue.Service.register] kafka[%s] of consumer[%s] has no task to route, check yaml[TaskRoute]",
					kafkaConf.Topic, routeKey))
				return
			}
			taskTemp, ok := sel.taskFunc[funcName]
			if !ok {
				// 指定的task类型没有注册
				logx.Errorf(fmt.Sprintf("[queue.Service.register] func{%s} not register, check [gen.go]", funcName))
				return
			}
			task := taskTemp(kafkaConf.Topic, sel)
			sel.tasks[funcName] = task
			kafkaConf.Telemetry = sel.telemetry
			consumer := kqkafka.NewConsumer(kafkaConf, task.Run)
			group.Add(consumer)
			logx.Infof("[queue.Service.register] reg func:%s, routeKey:%s", funcName, routeKey)
		}).Catch(func(e exception.Exception) {
			logx.Errorf("[queue.Service.register][ERROR] err:%s", e)
		}).Finally(func() {

		})
	}
}

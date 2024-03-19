//@File     : logic.go
//@Time     : 2024/3/18
//@Auther   : Kaishin

package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

func NewService(group *service.ServiceGroup, options ...OptionService) IService {
	svc := createSvc()
	for _, option := range options {
		option(svc)
	}
	svc.register()
	group.Add(svc)
	return svc
}

func (sel *Service) Gen(key string, t FuncTask) {
	sel.tasks[key] = t
	logx.Infof("[cron.Service.Gen] task:%s", key)
}

func (sel *Service) Stop() {
	sel.cron.Stop()
}

func (sel *Service) Start() {
	sel.cron.Start()
}

func createSvc() *Service {
	svc := new(Service)
	svc.cron = cron.New()
	svc.tasks = make(map[string]FuncTask)
	return svc
}

func (sel *Service) register() {
	// 获取注册函数
	funcRegister := sel.funcRegister
	if funcRegister == nil {
		logx.Errorf("[cron.Service.register] funcRegister is nil, check [gen.go]")
		return
	}

	// 注册任务
	funcRegister(sel)

	// 启动任务
	for _, conf := range sel.confList {
		task, ok := sel.tasks[conf.TaskType]
		if !ok {
			logx.Errorf("[cron.Service.register] 【%s】 taskType not exist", conf.TaskType)
			continue
		}
		_, err := sel.cron.AddJob(conf.Interval, task(conf.Params))
		if err != nil {
			logx.Errorf("[cron.Service.register] 【%s】 add task failed, params:%v, err:%v",
				conf.TaskType, conf.Params, err)
			continue
		}
		logx.Infof("[cron.Service.register] taskType:%s, conf:%v", conf.TaskType, conf)
	}
}

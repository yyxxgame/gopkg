//@File     : cron_test.go
//@Time     : 2024/3/18
//@Auther   : Kaishin

package cron

import (
	"github.com/mitchellh/mapstructure"
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"testing"
)

type (
	Template struct {
		Value int    `json:"value"`
		Str   string `json:"str"`
	}
)

func NewTemplate(params map[string]interface{}) ITask {
	var work = Template{}
	err := mapstructure.Decode(params, &work)
	if err != nil {
		logx.Errorf("[NewTemplate] error:%v, params:%v", err, params)
		return nil
	}
	logx.Infof("[NewTemplate] params:%v, addr:%p", params, &work)
	return &work
}

func (sel *Template) Run() {
	logx.Info(sel.Str, sel.Value)
}

func Gen(svc IService) {
	svc.Gen("Template", NewTemplate)
}

func TestTask(t *testing.T) {
	group := service.NewServiceGroup()
	defer group.Stop()

	confList := []ConfTask{
		{
			Name:     "Template",
			TaskType: "Template",
			Interval: "*/1 * * * * *",
			Params: map[string]interface{}{
				"str":   "hello",
				"value": 1,
			},
		},
	}

	svc := NewService(
		group,
		WithFuncRegister(Gen),
		WithConfList(confList),
		WithCronOptions(
			cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
			cron.WithSeconds(),
		),
	)
	defer svc.Stop()
}

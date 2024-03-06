//@File     : task_test.go
//@Time     : 2024/3/6
//@Auther   : Kaishin

package core

import (
	"context"
	"encoding/json"
	"github.com/yyxxgame/gopkg/exception"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/trace"
	"testing"
)

type (
	ServerCtx struct {
		KafkaConf []kq.KqConf
		TaskRoute map[string]string
		Telemetry trace.Config
	}

	TaskTemp struct {
		BaseTask
	}
)

func (sel *TaskTemp) Run(ctx context.Context, k, v string) error {
	var msg map[string]interface{}
	var err error
	err = json.Unmarshal([]byte(v), &msg)
	if err != nil {
		logx.Error(err)
		return err
	}

	exception.Try(func() {
		// todo do your business here
		logx.Infof("[Template.Run]  %v", msg)
	}).Catch(func(e exception.Exception) {
		//
		err = e.(error)
	}).Finally(func() {
		//
	})

	return err
}

func NewTaskTemplate(topic string, serverCtx ITaskServerCtx) ITask {
	obj := &TaskTemp{}
	obj.BaseTask.Topic = topic
	// obj.ServerCtx = serverCtx.Object().(*svc.ServiceContext) // todo use in business code
	return obj
}

func Register(factory ITaskFactory) {
	factory.Gen("test", NewTaskTemplate)
}

func TestTask(t *testing.T) {
	// defines
	ctx := ServerCtx{}
	group := service.NewServiceGroup()
	defer group.Stop()

	// register tasks
	defer StopTasks()
	RegisterTasks(
		NewTaskServerCtx(
			WithTaskRoute(ctx.TaskRoute),
			WithKafkaConf(ctx.KafkaConf),
			WithTelemetry(ctx.Telemetry),
			WithFuncRegister(Register),
		),
		group,
	)
}

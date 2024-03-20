//@File     : task_test.go
//@Time     : 2024/3/6
//@Auther   : Kaishin

package queue

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
	ServiceContext struct {
		KafkaConf []kq.KqConf
		TaskRoute map[string]string
		Telemetry trace.Config
	}

	TaskTemp struct {
		BaseTask
		serverCtx *ServiceContext
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

func (sel *TaskTemp) Stop() {
	logx.Info("TaskTemplate Stop")
}

func NewTaskTemplate(topic string, serverCtx IService) ITask {
	obj := &TaskTemp{}
	obj.BaseTask.Topic = topic
	obj.serverCtx = serverCtx.Object().(*ServiceContext) // todo use in business code
	return obj
}

func Register(svr IService) {
	svr.Gen("test", NewTaskTemplate)
}

func TestTask(t *testing.T) {
	// defines
	ctx := &ServiceContext{
		TaskRoute: map[string]string{"topic_name.group_name": "test"},
		KafkaConf: []kq.KqConf{
			{
				Brokers:    []string{"127.0.0.1:9092"},
				Group:      "group_name",
				Topic:      "topic_name",
				Processors: 1,
				Consumers:  1,
			},
		},
	}
	group := service.NewServiceGroup()
	defer group.Stop()

	// register tasks
	taskServer := NewService(
		group,
		WithTaskRoute(ctx.TaskRoute),
		WithKafkaConf(ctx.KafkaConf),
		WithTelemetry(ctx.Telemetry),
		WithFuncRegister(Register),
		WithObject(ctx),
	)
	defer taskServer.Stop()
}

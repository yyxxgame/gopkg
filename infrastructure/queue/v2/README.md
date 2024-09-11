* 服务配置文件

```yaml
...

QueueTaskConf:
  Brokers:
    - 127.0.0.1:9092
    - 127.0.0.1:9093
    - 127.0.0.1:9094
  Jobs:
    - Name: temp_job
      Topic: "temp-job-topic"
      Params:
        key1: "value1"
        key2: "value2"
      Enable: true
...

```


* `temp_job.go`文件

需要实现IQueueJob接口

```go
package your_package_name

import (
	"context"
	queuev2 "github.com/yyxxgame/gopkg/infrastructure/queue/v2"
)

type TempJob struct {
	// go-zero服务的svcCtx
	svcCtx *svc.ServiceContext
}

func NewTempJob(svcCtx *svc.ServiceContext) queuev2.IQueueJob {
	job := &TempJob{
		svcCtx: svcCtx,
	}

	return job
}
func (job *TempJob) OnConsume(ctx context.Context, payload string, params map[string]any) error {
	//TODO on exec temp job logic
	return nil
}

func (job *TempJob) Named() string {
	return "temp_job"
}

```


* 服务入口

```go
package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	queuev2 "github.com/yyxxgame/gopkg/infrastructure/queue/v2"
)

var configFile = flag.String("f", "etc/queue.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	group := service.NewServiceGroup()
	defer group.Stop()

	apiServer := rest.MustNewServer(c.RestConf)
	handler.RegisterHandlers(apiServer, ctx)
	group.Add(apiServer)

	queueTaskController := queuev2.NewQueueTaskController(c.QueueTaskConf)
	// 注册你的任务
	queueTaskController.RegisterJobs(
		job.NewTempJob(ctx),
	)
	group.Add(queueTaskController)

	fmt.Printf("Starting %s server at %s:%d...\n", c.Name, c.Host, c.Port)
	group.Start()
}

```
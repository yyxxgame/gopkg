* 服务配置文件

```yaml
...

CronTaskConf:
  Jobs:
    - Name: temp_job
      Expression: "*/5 * * * * ?"
      Params:
        key1: "value1"
        key2: "value2"
      Enable: true
...

```


* `temp_job.go`文件

需要实现ICronJob接口

```go
package your_package_name

type TempJob struct {
	// go-zero服务的svcCtx
	svcCtx *svc.ServiceContext
}

func NewTempJob(svcCtx *svc.ServiceContext) pkgcron.ICronJob {
	job := &TempJob{
		svcCtx: svcCtx,
	}

	return job
}
func (job *TempJob) OnExec(ctx context.Context, params map[string]any) error {
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
)

var configFile = flag.String("f", "etc/cron.yaml", "the config file")

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

	cronTaskController := pkgcron.NewCronTaskController(c.CronTaskConf)
	// 注册你的任务
	cronTaskController.RegisterJobs(
		job.NewTempJob(ctx),
	)
	group.Add(cronTaskController)

	fmt.Printf("Starting %s server at %s:%d...\n", c.Name, c.Host, c.Port)
	group.Start()
}

```
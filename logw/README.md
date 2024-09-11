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

	// 需要保证在所有服务初始化后初始化并注入 log.writer，不然会被 go-zero 内置的 writer 替换
	// 请看 go-zero 中 logx.AddWriter() 实现
	logx.AddGlobalFields(logx.LogField{
		Key:   "server_name",
		Value: c.Name,
	})
	p := saramakafka.NewProducer(c.LogKafkaConf.Brokers)
	w := logx.NewWriter(logw.MustNewKafkaWriter(p, c.LogKafkaConf.Topic))
	logx.SetWriter(w)

	fmt.Printf("Starting %s server at %s:%d...\n", c.Name, c.Host, c.Port)
	group.Start()
}

```
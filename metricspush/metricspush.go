//@File     metricspush.go
//@Time     2024/5/10
//@Author   #Suyghur,

package metricspush

import (
	"github.com/yyxxgame/gopkg/prompusher"
	"github.com/zeromicro/go-zero/core/proc"
)

var pusher prompusher.IPromMetricsPusher

type PushGateWayConfig struct {
	Addr     string `json:",optional"` // nolint
	Project  string `json:",optional"` // nolint
	Job      string `json:",optional"` // nolint
	Interval int64  `json:",optional"` // nolint
}

// InitMetricsPushService
// This func may be removed in the future, use prompusher.MustNewPromMetricsPusher instead
// Deprecated
func InitMetricsPushService(config PushGateWayConfig) {
	c := prompusher.PromPushConf{
		Url:         config.Addr,
		JobName:     config.Job,
		ProjectName: config.Project,
		Interval:    int(config.Interval),
	}
	pusher = prompusher.MustNewPromMetricsPusher(c)
	pusher.Start()

	proc.AddShutdownListener(func() {
		pusher.Stop()
	})
}

package prompusher

import (
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	prompush "github.com/prometheus/client_golang/prometheus/push"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/zeromicro/go-zero/core/logx"
	gozerometric "github.com/zeromicro/go-zero/core/metric"
	gozerostat "github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/timex"
)

var metricCpuUsage = gozerometric.NewGaugeVec(&gozerometric.GaugeVecOpts{
	Namespace: "process",
	Subsystem: "cpu",
	Name:      "usage",
	Help:      "process cpu usage.",
	Labels:    []string{},
})

type (
	PromPushConf struct {
		Url          string
		JobName      string
		ProjectName  string `json:",optional"` // nolint
		InstanceName string `json:",optional"` // nolint
		Interval     int    `json:",optional"` // nolint
	}

	IPromMetricsPusher interface {
		Start()
		Stop()
	}

	pusher struct {
		*prompush.Pusher
		ticker timex.Ticker
		done   *syncx.DoneChan

		enableRemoveOldMetrics bool
	}

	Option func(p *pusher)
)

func MustNewPromMetricsPusher(c PromPushConf, opts ...Option) IPromMetricsPusher {
	if c.Url == "" || c.JobName == "" {
		return nil
	}

	p := &pusher{
		Pusher: prompush.New(c.Url, c.JobName).Gatherer(prom.DefaultGatherer),
		ticker: timex.NewTicker(time.Duration(c.Interval) * time.Second),
		done:   syncx.NewDoneChan(),
	}

	if c.ProjectName != "" {
		p.Grouping("project", c.ProjectName)
	}

	if c.InstanceName != "" {
		p.Grouping("instance", c.InstanceName)
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.enableRemoveOldMetrics {
		_ = p.Delete()
	}

	return p
}

func (p *pusher) Start() {
	gopool.Go(func() {
		for {
			select {
			case <-p.ticker.Chan():
				// added CPU usage usage metrics
				metricCpuUsage.Set(float64(gozerostat.CpuUsage()))

				if err := p.Push(); err != nil {
					logx.Errorf("[PROM-PUSHER]: push metrics to pushgateway on error: %v", err)
				}
			case <-p.done.Done():
				p.ticker.Stop()
				return
			}
		}
	})
}

func (p *pusher) Stop() {
	p.done.Close()
}

// WithRemoveOldMetrics remove old metrics in pushgateway when startup.
func WithRemoveOldMetrics(enable bool) Option {
	return func(p *pusher) {
		p.enableRemoveOldMetrics = enable
	}
}

package prompusher

import (
	"regexp"
	"time"

	gozerostat "github.com/zeromicro/go-zero/core/stat"

	prom "github.com/prometheus/client_golang/prometheus"
	prompush "github.com/prometheus/client_golang/prometheus/push"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/zeromicro/go-zero/core/logx"
	gozerometric "github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/sysx"
	"github.com/zeromicro/go-zero/core/timex"
)

var (
	metricCpuUsage = gozerometric.NewGaugeVec(&gozerometric.GaugeVecOpts{
		Namespace: "process",
		Subsystem: "cpu",
		Name:      "usage",
		Help:      "process cpu usage.",
		Labels:    []string{},
	})
	matcher = regexp.MustCompile(`go_goroutines\{instance="([^"]*)",job="([^"]*)",project="([^"]*)"\}`)
)

type (
	PromPushConf struct {
		Url         string
		JobName     string
		ProjectName string
		Interval    int `json:",optional"` // nolint
	}

	IPromMetricsPusher interface {
		Start()
		Stop()
	}

	pusher struct {
		*prompush.Pusher

		ticker                       timex.Ticker
		done                         *syncx.DoneChan
		instanceName                 string
		enableCollectCpuUsageMetrics bool
	}

	pusherInfoPair struct {
		jobName      string
		projectName  string
		instanceName string
	}

	Option func(p *pusher)
)

func MustNewPromMetricsPusher(c PromPushConf, opts ...Option) IPromMetricsPusher {
	if c.Url == "" || c.JobName == "" || c.ProjectName == "" {
		return nil
	}

	p := &pusher{
		ticker:                       timex.NewTicker(time.Duration(c.Interval) * time.Second),
		done:                         syncx.NewDoneChan(),
		instanceName:                 sysx.Hostname(),
		enableCollectCpuUsageMetrics: true,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.Pusher = prompush.New(c.Url, c.JobName).
		Gatherer(prom.DefaultGatherer).
		Grouping("project", c.ProjectName).
		Grouping("instance", p.instanceName)

	return p
}

func (p *pusher) Start() {
	gopool.Go(func() {
		for {
			select {
			case <-p.ticker.Chan():
				// added CPU usage usage metrics
				if p.enableCollectCpuUsageMetrics {
					metricCpuUsage.Set(float64(gozerostat.CpuUsage()))
				}

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
	_ = p.Delete()
}

// WithInstanceName set pusher instance name.
func WithInstanceName(instanceName string) Option {
	return func(p *pusher) {
		p.instanceName = instanceName
	}
}

func WithCollectCpuUsageMetrics(enable bool) Option {
	return func(p *pusher) {
		p.enableCollectCpuUsageMetrics = enable
	}
}

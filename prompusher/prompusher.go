package prompusher

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	prompush "github.com/prometheus/client_golang/prometheus/push"
	"github.com/yyxxgame/gopkg/syncx/gopool"
	"github.com/zeromicro/go-zero/core/fx"
	"github.com/zeromicro/go-zero/core/logx"
	gozerometric "github.com/zeromicro/go-zero/core/metric"
	gozerostat "github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/sysx"
	"github.com/zeromicro/go-zero/core/timex"
	"github.com/zeromicro/go-zero/rest/httpc"
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
		enableRemoveOldMetrics       bool
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
		enableRemoveOldMetrics:       true,
		enableCollectCpuUsageMetrics: true,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.Pusher = prompush.New(c.Url, c.JobName).
		Gatherer(prom.DefaultGatherer).
		Grouping("project", c.ProjectName).
		Grouping("instance", p.instanceName)

	if p.enableRemoveOldMetrics {
		p.removeOldMetrics(c.Url)
	}

	return p
}

func (p *pusher) removeOldMetrics(url string) {
	gopool.Go(func() {
		resp, _ := httpc.Do(context.Background(), http.MethodGet, url, nil)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		fx.From(func(source chan<- any) {
			for scanner.Scan() {
				line := scanner.Text()
				matches := matcher.FindStringSubmatch(line)
				if len(matches) < 1 {
					continue
				}

				instance, job, project := matches[1], matches[2], matches[3]
				source <- &pusherInfoPair{
					jobName:      job,
					projectName:  project,
					instanceName: instance,
				}
			}
		}).Parallel(func(item any) {
			pair := item.(*pusherInfoPair)
			req, _ := http.NewRequestWithContext(
				context.Background(),
				http.MethodDelete,
				fmt.Sprintf("%s/job/%s/instance/%s/project/%s", url, pair.jobName, pair.instanceName, pair.projectName),
				nil,
			)
			_, _ = httpc.DoRequest(req)
		})
	})
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

// WithInstanceName set pusher instance name.
func WithInstanceName(instanceName string) Option {
	return func(p *pusher) {
		p.instanceName = instanceName
	}
}

// WithRemoveOldMetrics remove old metrics in pushgateway when startup.
func WithRemoveOldMetrics(enable bool) Option {
	return func(p *pusher) {
		p.enableRemoveOldMetrics = enable
	}
}

func WithCollectCpuUsageMetrics(enable bool) Option {
	return func(p *pusher) {
		p.enableCollectCpuUsageMetrics = enable
	}
}

package metricspush

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/threading"
)

type (
	PushGateWayConfig struct {
		Addr     string `json:",optional"` // nolint
		Project  string `json:",optional"` // nolint
		Job      string `json:",optional"` // nolint
		Interval int64  `json:",optional"` // nolint
	}
)

var metricCpuUsageGaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
	Namespace: "process",
	Subsystem: "cpu",
	Name:      "usage",
	Help:      "process cpu usage.",
	Labels:    []string{},
})

func InitMetricsPushService(config PushGateWayConfig) {
	if len(config.Addr) > 0 {
		threading.GoSafe(func() {
			url := fmt.Sprintf("http://%s/metrics", config.Addr)
			httpClient := &http.Client{Timeout: time.Second * 2}
			// Delete Past Instance
			re, _ := regexp.Compile(`go_goroutines\{instance="([^"]*)",job="([^"]*)",project="([^"]*)"\}`)
			resp, _ := http.Get(url)
			scanner := bufio.NewScanner(resp.Body)
			var isFind bool
			for scanner.Scan() {
				line := scanner.Text()
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					isFind = true
					instance, job, project := matches[1], matches[2], matches[3]
					if job == config.Job && project == config.Project {
						metricsDeleteReq, err := http.NewRequest(
							http.MethodDelete,
							fmt.Sprintf("%s/job/%s/instance/%s/project/%s", url, job, instance, project),
							nil,
						)
						if err == nil {
							_, _ = httpClient.Do(metricsDeleteReq)
						}
					}
				} else {
					if isFind {
						break
					}
				}
			}
			_ = resp.Body.Close()
			// Push To PushGateWay
			metricPushTicker := time.NewTicker(time.Second * time.Duration(config.Interval))
			defer metricPushTicker.Stop()
			for { // nolint
				select {
				case <-metricPushTicker.C:
					// CPU usage usage metrics
					metricCpuUsageGaugeVec.Set(float64(stat.CpuUsage()))
					mfs, mfsDone, err := prometheus.ToTransactionalGatherer(prometheus.DefaultGatherer).Gather()
					var mfsBuffer bytes.Buffer
					enc := expfmt.NewEncoder(&mfsBuffer, expfmt.FmtText)
					for _, mf := range mfs {
						err = enc.Encode(mf)
					}
					if err == nil {
						hostName, _ := os.Hostname()
						metricsPushReq, err := http.NewRequest(
							http.MethodPost,
							fmt.Sprintf("%s/job/%s/instance/%s/project/%s", url, config.Job, hostName, config.Project),
							&mfsBuffer,
						)
						if err == nil {
							_, _ = httpClient.Do(metricsPushReq)
						}
					}
					mfsDone()
				}
			}
		})
	}
}

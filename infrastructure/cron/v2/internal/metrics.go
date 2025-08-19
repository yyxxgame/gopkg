//@File     metrics.go
//@Time     2024/8/13
//@Author   #Suyghur,

package internal

import "github.com/zeromicro/go-zero/core/metric"

const namespace = "cron_task_controller"

var (
	MetricExecJobsDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "exec_jobs",
		Name:      "duration_ms",
		Help:      "cron task controller exec job duration(ms).",
		Labels:    []string{"cron_job_name"},
		Buckets:   []float64{100, 250, 500, 1000, 2000, 5000, 10000, 15000, 20000, 30000, 60000},
	})
	MetricExecJobsErr = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "exec_jobs",
		Name:      "error_total",
		Help:      "cron task controller exec job error count.",
		Labels:    []string{"cron_job_name"},
	})
	MetricExecJobsLastTimestamp = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: namespace,
		Subsystem: "exec_jobs",
		Name:      "last_timestamp",
		Help:      "cron task controller exec job last timestamp.",
		Labels:    []string{"cron_job_name"},
	})
)

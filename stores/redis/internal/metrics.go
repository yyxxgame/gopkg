//@File     metrics.go
//@Time     2025/5/28
//@Author   #Suyghur,

package internal

import (
	"fmt"
	"sync"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/prometheus/client_golang/prometheus"
	v9rds "github.com/redis/go-redis/v9"
	gozerometric "github.com/zeromicro/go-zero/core/metric"
)

const namespace = "go_redis_client"

var (
	metricReqDur = gozerometric.NewHistogramVec(&gozerometric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "redis client requests duration(ms).",
		Labels:    []string{"command"},
		Buckets:   []float64{0.25, 0.5, 1, 1.5, 2, 3, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000, 10000, 15000},
	})
	metricReqErr = gozerometric.NewCounterVec(&gozerometric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "error_total",
		Help:      "redis client requests error count.",
		Labels:    []string{"command", "error"},
	})
	metricSlowCount = gozerometric.NewCounterVec(&gozerometric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "slow_total",
		Help:      "redis client requests slow count.",
		Labels:    []string{"command"},
	})

	connLabels                         = []string{"host", "db"}
	ConnCollector                      = newCollector()
	_             prometheus.Collector = (*collector)(nil)
)

type (
	StatGetter struct {
		Host      string
		DB        int
		PoolSize  int
		PoolStats func() *v9rds.PoolStats
	}

	// collector collects statistics from a redis client.
	// It implements the prometheus.Collector interface.
	collector struct {
		hitDesc     *prometheus.Desc
		missDesc    *prometheus.Desc
		timeoutDesc *prometheus.Desc
		totalDesc   *prometheus.Desc
		idleDesc    *prometheus.Desc
		staleDesc   *prometheus.Desc
		maxDesc     *prometheus.Desc

		uniqueClients map[string]*StatGetter
		lock          sync.Mutex
	}
)

func newCollector() *collector {
	c := &collector{
		hitDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_hit_total"),
			"Number of times a connection was found in the pool",
			connLabels, nil,
		),
		missDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_miss_total"),
			"Number of times a connection was not found in the pool",
			connLabels, nil,
		),
		timeoutDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_timeout_total"),
			"Number of times a timeout occurred when looking for a connection in the pool",
			connLabels, nil,
		),
		totalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_conn_total_current"),
			"Current number of connections in the pool",
			connLabels, nil,
		),
		idleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_conn_idle_current"),
			"Current number of idle connections in the pool",
			connLabels, nil,
		),
		staleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_conn_stale_total"),
			"Number of times a connection was removed from the pool because it was stale",
			connLabels, nil,
		),
		maxDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "pool_conn_max"),
			"Max number of connections in the pool",
			connLabels, nil,
		),
		uniqueClients: make(map[string]*StatGetter),
	}

	prometheus.MustRegister(c)

	return c
}

// Describe implements the prometheus.Collector interface.
func (c *collector) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.hitDesc
	descs <- c.missDesc
	descs <- c.timeoutDesc
	descs <- c.totalDesc
	descs <- c.idleDesc
	descs <- c.staleDesc
	descs <- c.maxDesc
}

// Collect implements the prometheus.Collector interface.
func (c *collector) Collect(metrics chan<- prometheus.Metric) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, client := range c.uniqueClients {
		host, db := client.Host, convertor.ToString(client.DB)
		stats := client.PoolStats()

		metrics <- prometheus.MustNewConstMetric(
			c.hitDesc,
			prometheus.CounterValue,
			float64(stats.Hits),
			host,
			db,
		)
		metrics <- prometheus.MustNewConstMetric(
			c.missDesc,
			prometheus.CounterValue,
			float64(stats.Misses),
			host,
			db,
		)
		metrics <- prometheus.MustNewConstMetric(
			c.timeoutDesc,
			prometheus.CounterValue,
			float64(stats.Timeouts),
			host,
			db,
		)
		metrics <- prometheus.MustNewConstMetric(
			c.totalDesc,
			prometheus.GaugeValue,
			float64(stats.TotalConns),
			host,
			db,
		)
		metrics <- prometheus.MustNewConstMetric(
			c.idleDesc,
			prometheus.GaugeValue,
			float64(stats.IdleConns),
			host,
			db,
		)
		metrics <- prometheus.MustNewConstMetric(
			c.staleDesc,
			prometheus.CounterValue,
			float64(stats.StaleConns),
			host,
			db,
		)
		metrics <- prometheus.MustNewConstMetric(
			c.maxDesc,
			prometheus.CounterValue,
			float64(client.PoolSize),
			host,
			db,
		)
	}
}

func (c *collector) RegisterClient(client *StatGetter) {
	c.lock.Lock()
	defer c.lock.Unlock()

	uniqueClientId := fmt.Sprintf("%s_%d", client.Host, client.DB)
	if maputil.HasKey(c.uniqueClients, uniqueClientId) {
		return
	}

	c.uniqueClients[uniqueClientId] = client
}

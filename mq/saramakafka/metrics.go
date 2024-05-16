//@File     metrics.go
//@Time     2024/5/16
//@Author   #Suyghur,

package saramakafka

import gozerometric "github.com/zeromicro/go-zero/core/metric"

const namespace = "saramakafka"

var (
	metricProduceDur = gozerometric.NewHistogramVec(&gozerometric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "producer",
		Name:      "duration_ms",
		Help:      "kafka producer client produce duration(ms).",
		Labels:    []string{"topic"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 750, 1000, 2000},
	})

	metricProduceErr = gozerometric.NewCounterVec(&gozerometric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "producer",
		Name:      "error_total",
		Help:      "kafka producer client produce error count.",
		Labels:    []string{"topic"},
	})

	metricConsumeDur = gozerometric.NewHistogramVec(&gozerometric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "consumer_group",
		Name:      "duration_ms",
		Help:      "kafka consumer group client consume duration(ms).",
		Labels:    []string{"topic", "group_id"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 750, 1000, 2000},
	})

	metricConsumeErr = gozerometric.NewCounterVec(&gozerometric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "consumer_group",
		Name:      "error_total",
		Help:      "kafka consumer group_client consumer error count.",
		Labels:    []string{"topic", "group_id"},
	})

	metricConsumerGroupLag = gozerometric.NewGaugeVec(&gozerometric.GaugeVecOpts{
		Namespace: namespace,
		Subsystem: "consumer_group",
		Name:      "lag",
		Help:      "current approximate lag of a consumer group at topic/partition",
		Labels:    []string{"topic", "group_id"},
	})

	metricConsumerGroupLagSum = gozerometric.NewGaugeVec(&gozerometric.GaugeVecOpts{
		Namespace: namespace,
		Subsystem: "consumer_group",
		Name:      "lag_sum",
		Help:      "current approximate lag of a consumer group at topic for all partitions",
		Labels:    []string{"topic", "group_id"},
	})
)

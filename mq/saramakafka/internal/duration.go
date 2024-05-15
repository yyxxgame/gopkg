//@File     duration.go
//@Time     2024/5/14
//@Author   #Suyghur,

package internal

import (
	"context"

	gozerometric "github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/timex"
	oteltrace "go.opentelemetry.io/otel/trace"
)

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
		Subsystem: "consumer",
		Name:      "duration_ms",
		Help:      "kafka consumer client consume duration(ms).",
		Labels:    []string{"topic"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 750, 1000, 2000},
	})
	metricConsumeErr = gozerometric.NewCounterVec(&gozerometric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "consumer",
		Name:      "error_total",
		Help:      "kafka consumer client consumer error count.",
		Labels:    []string{"topic"},
	})
)

type DurationHook struct {
	spanKind oteltrace.SpanKind
}

func NewDurationHook(spanKind oteltrace.SpanKind) *DurationHook {
	return &DurationHook{
		spanKind: spanKind,
	}
}

func (c *DurationHook) Handle(ctx context.Context, topic, key, payload string, next HookFunc) error {
	start := timex.Now()

	err := next(ctx, topic, key, payload)

	duration := timex.Since(start)

	switch c.spanKind {
	case oteltrace.SpanKindProducer:
		metricProduceDur.Observe(duration.Milliseconds(), topic)
		if err != nil {
			metricProduceErr.Inc(topic)
		}
	case oteltrace.SpanKindConsumer:
		metricConsumeDur.Observe(duration.Milliseconds(), topic)
		if err != nil {
			metricConsumeErr.Inc(topic)
		}
	default:
		//ignore
	}

	return err
}

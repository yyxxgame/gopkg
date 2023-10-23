//@File     metric.go
//@Time     2023/10/23
//@Author   #Suyghur,

package gormc

import (
	"github.com/zeromicro/go-zero/core/metric"
	"gorm.io/gorm"
	"strconv"
	"time"
)

const namespace = "gormc_client"

var (
	metricReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "gormc_duration_ms",
		Help:      "gormc requests duration(ms).",
		Labels:    []string{"table", "method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	})

	metricReqErr = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "gormc_error_total",
		Help:      "gormc requests error count.",
		Labels:    []string{"table", "method", "is_error"},
	})
)

type MetricPlugin struct {
}

func (p *MetricPlugin) Name() string {
	return namespace
}

func (p *MetricPlugin) Initialize(db *gorm.DB) error {
	if err := p.addCreateHooks(db); err != nil {
		return err
	}

	if err := p.addQueryHooks(db); err != nil {
		return err
	}
	if err := p.addUpdateHooks(db); err != nil {
		return err
	}
	if err := p.addDeleteHooks(db); err != nil {
		return err
	}
	if err := p.addRowHooks(db); err != nil {
		return err
	}
	if err := p.addRawHooks(db); err != nil {
		return err
	}
	return nil
}

func (p *MetricPlugin) addCreateHooks(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gormc:create_before").Register("gormc:create_before:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gormc:create_after").Register("gormc:create_after:metric", func(db *gorm.DB) {
		if ts, ok := db.InstanceGet("gormc:create_ts"); !ok {
			return
		} else {
			sec := ts.(int64)
			st := time.Unix(sec, 0)
			metricReqDur.Observe(time.Since(st).Milliseconds(), db.Statement.Table, "create")
			metricReqErr.Inc(db.Statement.Table, "create", strconv.FormatBool(db.Statement.Error != nil))
			//TODO get trace span?
		}
	}); err != nil {
		return err
	}
	return nil
}

func (p *MetricPlugin) addQueryHooks(db *gorm.DB) error {
	if err := db.Callback().Query().Before("gormc:query_before").Register("gormc:query_before:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gormc:query_after").Register("gormc:query_after:metric", func(db *gorm.DB) {
		if ts, ok := db.InstanceGet("gormc:create_ts"); !ok {
			return
		} else {
			sec := ts.(int64)
			st := time.Unix(sec, 0)
			metricReqDur.Observe(time.Since(st).Milliseconds(), db.Statement.Table, "query")
			metricReqErr.Inc(db.Statement.Table, "query", strconv.FormatBool(db.Statement.Error != nil))
			//TODO get trace span?
		}
	}); err != nil {
		return err
	}
	return nil
}

func (p *MetricPlugin) addUpdateHooks(db *gorm.DB) error {
	if err := db.Callback().Update().Before("gormc:update_before").Register("gormc:update_before:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gormc:update_after").Register("gormc:update_after:metric", func(db *gorm.DB) {
		if ts, ok := db.InstanceGet("gormc:create_ts"); !ok {
			return
		} else {
			sec := ts.(int64)
			st := time.Unix(sec, 0)
			metricReqDur.Observe(time.Since(st).Milliseconds(), db.Statement.Table, "update")
			metricReqErr.Inc(db.Statement.Table, "update", strconv.FormatBool(db.Statement.Error != nil))
			//TODO get trace span?
		}
	}); err != nil {
		return err
	}
	return nil
}

func (p *MetricPlugin) addDeleteHooks(db *gorm.DB) error {
	if err := db.Callback().Delete().Before("gormc:delete_before").Register("gormc:delete_before:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gormc:delete_after").Register("gormc:delete_after:metric", func(db *gorm.DB) {
		if ts, ok := db.InstanceGet("gormc:create_ts"); !ok {
			return
		} else {
			sec := ts.(int64)
			st := time.Unix(sec, 0)
			metricReqDur.Observe(time.Since(st).Milliseconds(), db.Statement.Table, "delete")
			metricReqErr.Inc(db.Statement.Table, "delete", strconv.FormatBool(db.Statement.Error != nil))
			//TODO get trace span?
		}
	}); err != nil {
		return err
	}
	return nil
}

func (p *MetricPlugin) addRowHooks(db *gorm.DB) error {
	if err := db.Callback().Row().Before("gormc:row_before").Register("gormc:row_before:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gormc:row_after").Register("gormc:row_after:metric", func(db *gorm.DB) {
		if ts, ok := db.InstanceGet("gormc:create_ts"); !ok {
			return
		} else {
			sec := ts.(int64)
			st := time.Unix(sec, 0)
			metricReqDur.Observe(time.Since(st).Milliseconds(), db.Statement.Table, "row")
			metricReqErr.Inc(db.Statement.Table, "row", strconv.FormatBool(db.Statement.Error != nil))
			//TODO get trace span?
		}
	}); err != nil {
		return err
	}
	return nil
}

func (p *MetricPlugin) addRawHooks(db *gorm.DB) error {
	if err := db.Callback().Raw().Before("gormc:raw_before").Register("gormc:raw_before:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gormc:raw_after").Register("gormc:raw_after:metric", func(db *gorm.DB) {
		if ts, ok := db.InstanceGet("gormc:create_ts"); !ok {
			return
		} else {
			sec := ts.(int64)
			st := time.Unix(sec, 0)
			metricReqDur.Observe(time.Since(st).Milliseconds(), db.Statement.Table, "raw")
			metricReqErr.Inc(db.Statement.Table, "raw", strconv.FormatBool(db.Statement.Error != nil))
			//TODO get trace span?
		}
	}); err != nil {
		return err
	}
	return nil
}

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

type metricPlugin struct {
}

func newMetricPlugin() *metricPlugin {
	return &metricPlugin{}
}

func (p *metricPlugin) Name() string {
	return namespace
}

func (p *metricPlugin) Initialize(db *gorm.DB) error {
	if err := p.addCreateHooks(db); err != nil {
		return err
	}

	if err := p.addQueryHooks(db); err != nil {
		return err
	}

	if err := db.Callback().Query().After("gormc:queryAfter").Register("gormc:queryAfter:metric", func(db *gorm.DB) {

	}); err != nil {
		return err
	}

	if err := db.Callback().Update().After("gormc:updateAfter").Register("gormc:updateAfter:metric", func(db *gorm.DB) {

	}); err != nil {
		return err
	}

	if err := db.Callback().Delete().After("gormc:deleteAfter").Register("gormc:deleteAfter:metric", func(db *gorm.DB) {

	}); err != nil {
		return err
	}

	if err := db.Callback().Row().After("gormc:rowAfter").Register("gormc:rowAfter:metric", func(db *gorm.DB) {

	}); err != nil {
		return err
	}

	if err := db.Callback().Raw().After("gormc:rawAfter").Register("gormc:rawAfter:metric", func(db *gorm.DB) {

	}); err != nil {
		return err
	}

	return nil
}

func (p *metricPlugin) addCreateHooks(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gormc:createBefore").Register("gormc:createBefore:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gormc:createAfter").Register("gormc:createAfter:metric", func(db *gorm.DB) {
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

func (p *metricPlugin) addQueryHooks(db *gorm.DB) error {
	if err := db.Callback().Query().Before("gormc:queryBefore").Register("gormc:queryBefore:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gormc:queryAfter").Register("gormc:queryAfter:metric", func(db *gorm.DB) {
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

func (p *metricPlugin) addUpdateHooks(db *gorm.DB) error {
	if err := db.Callback().Update().Before("gormc:updateBefore").Register("gormc:updateBefore:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gormc:updateAfter").Register("gormc:updateAfter:metric", func(db *gorm.DB) {
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

func (p *metricPlugin) addDeleteHooks(db *gorm.DB) error {
	if err := db.Callback().Delete().Before("gormc:deleteBefore").Register("gormc:deleteBefore:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gormc:deleteAfter").Register("gormc:deleteAfter:metric", func(db *gorm.DB) {
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

func (p *metricPlugin) addRowHooks(db *gorm.DB) error {
	if err := db.Callback().Row().Before("gormc:rowBefore").Register("gormc:rowBefore:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gormc:rowAfter").Register("gormc:rowAfter:metric", func(db *gorm.DB) {
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

func (p *metricPlugin) addRawHooks(db *gorm.DB) error {
	if err := db.Callback().Raw().Before("gormc:rawBefore").Register("gormc:rawBefore:metric", func(db *gorm.DB) {
		ts := time.Now().Unix()
		db.InstanceSet("gormc:create_ts", ts)
		//TODO add trace span?
	}); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gormc:rawAfter").Register("gormc:rawAfter:metric", func(db *gorm.DB) {
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

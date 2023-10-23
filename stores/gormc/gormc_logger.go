//@File     gorm_logger.go
//@Time     2023/09/14
//@Author   #Suyghur,

package gormc

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

const (
	// DevMode means development mode.
	DevMode = "dev"
	// ProMode means production mode.
	ProMode = "pro"
)

var _ logger.Interface = (*GormLogger)(nil)

type GormLogger struct {
	Mode          string
	SlowThreshold time.Duration
}

func NewGormLogger(mode string) *GormLogger {
	return &GormLogger{
		Mode:          mode,
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{
		Mode:          l.Mode,
		SlowThreshold: l.SlowThreshold,
	}
}

func (l *GormLogger) Info(ctx context.Context, format string, v ...interface{}) {
	logx.WithContext(ctx).Infof(format, v)
}

func (l *GormLogger) Warn(ctx context.Context, format string, v ...interface{}) {
	logx.WithContext(ctx).Errorf(format, v)
}

func (l *GormLogger) Error(ctx context.Context, format string, v ...interface{}) {
	logx.WithContext(ctx).Errorf(format, v)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 获取运行时间
	elapsed := time.Since(begin)
	// 获取 SQL 语句和返回条数
	sql, rows := fc()
	// 通用字段
	logFields := []logx.LogField{
		logx.Field("sql", sql),
		logx.Field("time", microsecondsStr(elapsed)),
		logx.Field("rows", rows),
	}
	// gorm 错误
	if err != nil {
		// 记录未找到的错误使用 warning 等级
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logx.WithContext(ctx).Infow("gorm record not found", logFields...)
		} else {
			// 其他错误使用 error 等级
			logFields = append(logFields, logx.Field("catch error", err))
			logx.WithContext(ctx).Errorw("gorm error", logFields...)
		}

		//// 其他错误使用 error 等级
		//logFields = append(logFields, logx.Field("catch error", err))
		//logx.WithContext(ctx).Errorw("gorm error sql", logFields...)
	}
	// 慢查询日志
	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		logx.WithContext(ctx).Sloww("gorm slow sql", logFields...)
	}
	// 非生产模式下，记录所有 SQL 请求
	if l.Mode != ProMode {
		logx.WithContext(ctx).Infow("gorm exec sql", logFields...)
	}
}

// microsecondsStr 将 time.Duration 类型（nano seconds 为单位）输出为小数点后 3 位的 ms （microsecond 毫秒，千分之一秒）
func microsecondsStr(elapsed time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
}

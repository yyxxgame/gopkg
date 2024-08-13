//@File     duration.go
//@Time     2024/8/13
//@Author   #Suyghur,

package internal

import (
	"context"

	"github.com/zeromicro/go-zero/core/timex"
)

func DurationHook(jobName string) Hook {
	return func(ctx context.Context, next func(ctx context.Context) error) error {
		start := timex.Now()

		err := next(ctx)

		duration := timex.Since(start)
		MetricExecJobsDur.Observe(duration.Milliseconds(), jobName)

		return err
	}
}

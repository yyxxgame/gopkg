//@File     job.go
//@Time     2024/8/12
//@Author   #Suyghur,

package cron

import (
	"context"

	"github.com/yyxxgame/gopkg/infrastructure/cron/v2/internal"
)

type (
	WrapperJob struct {
		cronJob   ICronJob
		params    map[string]any
		finalHook internal.Hook
	}
)

func (job *WrapperJob) Run() {
	ctx := context.Background()

	_ = job.finalHook(ctx, func(ctx context.Context) error {
		err := job.cronJob.OnExec(ctx, job.params)
		if err != nil {
			internal.MetricExecJobsErr.Inc(job.cronJob.Named())
		}

		return err
	})
}

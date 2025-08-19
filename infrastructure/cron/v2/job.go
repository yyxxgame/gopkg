//@File     job.go
//@Time     2024/8/12
//@Author   #Suyghur,

package v2

import (
	"context"
)

type (
	ICronJob interface {
		OnExec(ctx context.Context, params map[string]any) error
		Named() string
	}

	WrapperJob struct {
		cronJob   ICronJob
		params    map[string]any
		finalHook Hook
	}
)

func (job *WrapperJob) Run() {
	ctx := context.Background()

	_ = job.finalHook(ctx, job.cronJob, func(ctx context.Context, j ICronJob) error {
		return j.OnExec(ctx, job.params)
	})
}

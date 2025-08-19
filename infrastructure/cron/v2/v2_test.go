//@File     v2_test.go
//@Time     2025/8/19
//@Author   #Suyghur,

package v2

import (
	"context"
	"testing"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
)

var (
	fakeConf = CronTaskConf{
		Jobs: []CronJobConf{
			{
				Name:       "mock_job",
				Expression: "*/10 * * * * *",
				Params:     make(map[string]any),
				Enable:     true,
			},
		},
	}
)

type (
	mockJob struct {
	}

	mockHook struct {
	}
)

func (job *mockJob) OnExec(ctx context.Context, _ map[string]any) error {
	logx.WithContext(ctx).Infof("OnExec")
	return nil
}

func (job *mockJob) Named() string {
	return "mock_job"
}

func (h *mockHook) ExecHook() Hook {
	return func(ctx context.Context, job ICronJob, next func(ctx context.Context, job ICronJob) error) error {
		logx.WithContext(ctx).Infof("MockHook start...")

		err := next(ctx, job)

		logx.WithContext(ctx).Infof("MockHook end...")

		return err
	}
}

func TestHook(t *testing.T) {
	done := syncx.NewDoneChan()

	mh := &mockHook{}
	c := NewCronTaskController(
		fakeConf,
		WithHook(mh.ExecHook()),
		WithTracer(otel.GetTracerProvider().Tracer(gozerotrace.TraceName)))
	c.RegisterJobs(
		&mockJob{},
	)

	c.Start()

	<-done.Done()

	c.Stop()
}

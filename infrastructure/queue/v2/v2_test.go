//@File     v2_test.go
//@Time     2025/10/20
//@Author   #Suyghur,

package v2

import (
	"context"
	"fmt"
	"testing"

	"github.com/zeromicro/go-zero/core/syncx"
)

type (
	fakeJob struct{}
)

var (
	fakeQueueTaskConf = QueueTaskConf{
		Brokers: []string{"127.0.0.1:9092"},
		Jobs: []QueueJobConf{
			{
				Name:      "fake_job",
				Topic:     "test_queue_topic",
				WorkerNum: 3,
				Enable:    true,
			},
		},
	}
)

func (job *fakeJob) OnConsume(ctx context.Context, payload string, params map[string]any) error {
	fmt.Printf("payload: %s\n", payload)
	return nil
}

func (job *fakeJob) Named() string {
	return "fake_job"
}

func TestRegisterJobs(t *testing.T) {
	done := syncx.NewDoneChan()

	c := NewQueueTaskController(fakeQueueTaskConf)
	c.RegisterJobs(&fakeJob{})

	c.Start()

	<-done.Done()
}

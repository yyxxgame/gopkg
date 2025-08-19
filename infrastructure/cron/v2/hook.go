//@File     hook.go
//@Time     2025/8/19
//@Author   #Suyghur,

package v2

import "context"

type (
	Hook func(ctx context.Context, job ICronJob, next func(ctx context.Context, job ICronJob) error) error
)

func chainHooks(hooks ...Hook) Hook {
	if len(hooks) <= 0 {
		return nil
	}

	return func(ctx context.Context, job ICronJob, next func(ctx context.Context, job ICronJob) error) error {
		return hooks[0](ctx, job, getNext(hooks, 0, next))
	}
}

func getNext(hooks []Hook, index int, next func(ctx context.Context, job ICronJob) error) func(ctx context.Context, job ICronJob) error {
	if index == len(hooks)-1 {
		return next
	}

	return func(ctx context.Context, job ICronJob) error {
		return hooks[index+1](ctx, job, getNext(hooks, index+1, next))
	}
}

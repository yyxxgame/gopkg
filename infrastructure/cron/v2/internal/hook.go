//@File     hook.go
//@Time     2024/8/12
//@Author   #Suyghur,

package internal

import "context"

type (
	Hook func(ctx context.Context, next func(ctx context.Context) error) error
)

func ChainHooks(hooks ...Hook) Hook {
	if len(hooks) <= 0 {
		return nil
	}
	return func(ctx context.Context, next func(ctx context.Context) error) error {
		return hooks[0](ctx, getNext(hooks, 0, next))
	}
}

func getNext(hooks []Hook, index int, final func(ctx context.Context) error) func(ctx context.Context) error {
	if index == len(hooks)-1 {
		return final
	}

	return func(ctx context.Context) error {
		return hooks[index+1](ctx, getNext(hooks, index+1, final))
	}
}

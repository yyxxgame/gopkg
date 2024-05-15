//@File     hook.go
//@Time     2024/5/15
//@Author   #Suyghur,

package internal

import (
	"context"
)

type (
	HookFunc func(ctx context.Context, topic, key, payload string) error

	Hook func(ctx context.Context, topic, key, payload string, next HookFunc) error
)

func ChainHooks(hooks ...Hook) Hook {
	if len(hooks) == 0 {
		return nil
	}
	return func(ctx context.Context, topic, key, payload string, next HookFunc) error {
		return hooks[0](ctx, topic, key, payload, getHookFunc(hooks, 0, next))
	}
}

func getHookFunc(hooks []Hook, index int, final HookFunc) HookFunc {
	if index == len(hooks)-1 {
		return final
	}

	return func(ctx context.Context, topic, key, payload string) error {
		return hooks[index+1](ctx, topic, key, payload, getHookFunc(hooks, index+1, final))
	}
}

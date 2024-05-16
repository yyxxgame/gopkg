//@File     hook.go
//@Time     2024/5/15
//@Author   #Suyghur,

package saramakafka

import "context"

type (
	hookFunc func(ctx context.Context, topic, key, payload string) error

	hook func(ctx context.Context, topic, key, payload string, next hookFunc) error
)

func chainHooks(hooks ...hook) hook {
	if len(hooks) == 0 {
		return nil
	}
	return func(ctx context.Context, topic, key, payload string, next hookFunc) error {
		return hooks[0](ctx, topic, key, payload, getHookFunc(hooks, 0, next))
	}
}

func getHookFunc(hooks []hook, index int, final hookFunc) hookFunc {
	if index == len(hooks)-1 {
		return final
	}

	return func(ctx context.Context, topic, key, payload string) error {
		return hooks[index+1](ctx, topic, key, payload, getHookFunc(hooks, index+1, final))
	}
}

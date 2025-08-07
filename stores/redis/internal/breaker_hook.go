//@File     breaker_hook.go
//@Time     2025/5/28
//@Author   #Suyghur,

package internal

import (
	"context"

	v9rds "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/breaker"
)

type BreakerHook struct {
	Brk breaker.Breaker
}

func (h *BreakerHook) DialHook(next v9rds.DialHook) v9rds.DialHook {
	return next
}

func (h *BreakerHook) ProcessHook(next v9rds.ProcessHook) v9rds.ProcessHook {
	return func(ctx context.Context, cmd v9rds.Cmder) error {
		if _, ok := ignoreCmds[cmd.Name()]; ok {
			return next(ctx, cmd)
		}

		return h.Brk.DoWithAcceptableCtx(ctx, func() error {
			return next(ctx, cmd)
		}, acceptable)
	}
}

func (h *BreakerHook) ProcessPipelineHook(next v9rds.ProcessPipelineHook) v9rds.ProcessPipelineHook {
	return func(ctx context.Context, cmds []v9rds.Cmder) error {
		for _, item := range cmds {
			if _, ok := ignoreCmds[item.Name()]; ok {
				return next(ctx, cmds)
			}
		}
		return h.Brk.DoWithAcceptableCtx(ctx, func() error {
			return next(ctx, cmds)
		}, acceptable)
	}
}

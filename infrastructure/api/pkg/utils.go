package pkg

import (
	"context"

	"github.com/duke-git/lancet/v2/convertor"
)

func GetIpFromContext(ctx context.Context) string {
	k := convertor.ToString(ctx.Value("__ipKey"))
	return convertor.ToString(ctx.Value(k))
}

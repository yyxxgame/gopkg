package pkg

import (
	"context"
)

func GetIpFromContext(ctx context.Context) string {
	k, _ := ctx.Value("__ipKey").(string)
	v, _ := ctx.Value(k).(string)
	return v
}

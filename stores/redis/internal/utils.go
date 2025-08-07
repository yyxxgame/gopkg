//@File     utils.go
//@Time     2025/2/18
//@Author   #Suyghur,

package internal

import (
	"context"

	v9rds "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/lang"
)

var (
	ignoreCmds = map[string]lang.PlaceholderType{
		"blpop":  {},
		"hello":  {},
		"client": {},
	}
)

func acceptable(err error) bool {
	return err == nil || errorx.In(err, v9rds.Nil, context.Canceled)
}

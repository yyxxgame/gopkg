//@File     : facade.go
//@Time     : 2024/3/6
//@Auther   : Kaishin

package queue

import (
	"context"
	"github.com/zeromicro/go-zero/core/service"
)

type (
	IService interface {
		service.Service
		Object() interface{}
		Gen(key string, t FuncTask)
	}

	ITask interface {
		Run(ctx context.Context, k, v string) error
		Stop()
	}
)

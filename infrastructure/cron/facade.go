//@File     : facade.go
//@Time     : 2024/3/18
//@Auther   : Kaishin

package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/service"
)

type (
	ITask interface {
		cron.Job
	}

	IService interface {
		service.Service
		Gen(string, FuncTask)
	}
)

//@File     : types.go
//@Time     : 2024/3/18
//@Auther   : Kaishin

package cron

import (
	"github.com/robfig/cron/v3"
)

type (
	ConfTask struct {
		Name     string
		TaskType string
		Interval string
		Params   map[string]interface{}
	}

	Service struct {
		IService
		cron         *cron.Cron
		tasks        map[string]FuncTask
		funcRegister FuncRegister
		confList     []ConfTask
	}

	FuncTask func(params map[string]interface{}) ITask

	FuncRegister func(svc IService)
)

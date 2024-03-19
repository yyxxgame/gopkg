//@File     : options.go
//@Time     : 2024/3/18
//@Auther   : Kaishin

package cron

import "github.com/robfig/cron/v3"

type (
	OptionService func(ctx *Service)
)

func WithFuncRegister(cb FuncRegister) OptionService {
	return func(o *Service) {
		o.funcRegister = cb
	}
}

func WithConfList(confList []ConfTask) OptionService {
	return func(o *Service) {
		o.confList = confList
	}
}

func WithCronOptions(opts ...cron.Option) OptionService {
	return func(o *Service) {
		for _, opt := range opts {
			opt(o.cron)
		}
	}
}

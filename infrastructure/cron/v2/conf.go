//@File     conf.go
//@Time     2024/8/10
//@Author   #Suyghur,

package cron

type (
	CronTaskConf struct {
		DurationInterceptor bool `json:",optional,default=true"`
		TraceInterceptor    bool `json:",optional,default=true"`
		Jobs                []CronJobConf
	}

	CronJobConf struct {
		Name       string
		Expression string
		Params     map[string]any
		Enable     bool `json:",optional,default=true"`
	}
)

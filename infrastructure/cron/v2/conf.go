//@File     conf.go
//@Time     2024/8/10
//@Author   #Suyghur,

package v2

type (
	CronTaskConf struct {
		Jobs []CronJobConf `json:",optional"`
	}

	CronJobConf struct {
		Name       string
		Expression string
		Params     map[string]any
		Enable     bool `json:",optional,default=true"`
	}
)

//@File     conf.go
//@Time     2024/9/11
//@Author   #Suyghur,

package v2

type (
	QueueTaskConf struct {
		Brokers  []string
		Username string `json:",optional"`
		Password string `json:",optional"`
		Jobs     []QueueJobConf
	}

	QueueJobConf struct {
		Name   string
		Topic  string
		Params map[string]any
		Enable bool `json:",optional,default=true"`
	}
)

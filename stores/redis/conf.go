//@File     conf.go
//@Time     2025/5/28
//@Author   #Suyghur,

package redis

type (
	RedisConf struct {
		Host     string
		Pass     string `json:",optional"`
		DB       int    `json:",optional"`
		NonBlock bool   `json:",default=true"`
	}
)

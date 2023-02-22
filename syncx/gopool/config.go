//@File     config.go
//@Time     2023/02/22
//@Author   #Suyghur,

package gopool

const (
	defaultScalaThreshold = 1
)

// Config is used to config pool.
type Config struct {
	// threshold for scale.
	// new goroutine is created if len(task chan) > ScaleThreshold.
	// defaults to defaultScalaThreshold.
	ScaleThreshold int32
}

// NewConfig creates a default Config.
func NewConfig() *Config {
	c := &Config{
		ScaleThreshold: defaultScalaThreshold,
	}
	return c
}

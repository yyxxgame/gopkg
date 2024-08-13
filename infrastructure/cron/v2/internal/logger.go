//@File     logger.go
//@Time     2024/7/16
//@Author   #Suyghur,

package internal

import "github.com/zeromicro/go-zero/core/logx"

var DefaultLogger = &cronTaskLogger{}

type cronTaskLogger struct{}

func (logger *cronTaskLogger) Info(msg string, keysAndValues ...any) {
	logx.Infof(msg, keysAndValues...)
}

func (logger *cronTaskLogger) Error(err error, msg string, keysAndValues ...any) {
	logx.Errorf(msg, keysAndValues...)
}

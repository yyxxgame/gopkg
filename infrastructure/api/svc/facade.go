package svc

import (
	"github.com/zeromicro/go-zero/rest"
)

type (
	IServiceContext interface {
		InitExtension(c ServiceConf)
		RegisterExtension(server *rest.Server)
	}
)

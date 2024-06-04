package svc

import (
	"net/http"

	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"
	"github.com/yyxxgame/gopkg/prompusher"
	"github.com/zeromicro/go-zero/rest"
)

type (
	ServiceConf struct {
		AccessLogDetails bool                    `json:",optional"` // nolint
		EncryptionKey    string                  `json:",optional"` // nolint
		ApiKey           string                  `json:",optional"` // nolint
		PromPushConf     prompusher.PromPushConf `json:",optional"` // nolint
	}

	ServiceContext struct {
		IServiceContext
		Config    ServiceConf
		RouterMap map[string]map[string]http.HandlerFunc
		Responder responder.IResponder

		GetIpMiddleware   rest.Middleware
		AuthMiddleware    rest.Middleware
		TransitMiddleware rest.Middleware
	}
)

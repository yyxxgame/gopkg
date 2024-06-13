package svc

import (
	"net/http"

	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"
	"github.com/yyxxgame/gopkg/prompusher"
	"github.com/zeromicro/go-zero/rest"
)

func (sel *ServiceContext) InitExtension(c ServiceConf) {
	sel.Config = c
	sel.initResponder()
	if len(sel.Config.PromPushConf.Url) > 0 {
		// PromMetricsPusher
		promPusher := prompusher.MustNewPromMetricsPusher(sel.Config.PromPushConf, prompusher.WithCleanupWhenShutdown(true))
		promPusher.Start()
	}
}

func (sel *ServiceContext) initResponder() {
	options := []responder.Option{responder.WithAccessLogDetails(sel.Config.AccessLogDetails)}
	if len(sel.Config.EncryptionKey) > 0 {
		options = append(
			options,
			responder.WithEncryptionHeader("Is-Decrypt"),
			responder.WithEncryptionKey(sel.Config.EncryptionKey),
		)
	}
	sel.Responder = responder.New(options...)
}

func (sel *ServiceContext) RegisterExtension(server *rest.Server) {
	if len(sel.Config.EncryptionKey) > 0 {
		sel.registerRouterMap(server)
	}
	sel.registerHeader(server)
}

func (sel *ServiceContext) registerHeader(server *rest.Server) {
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			request.Header.Set("__params", "")
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Headers", "*")
			writer.Header().Set("Access-Control-Allow-Methods", "*")
			next(writer, request)
		}
	})
}

func (sel *ServiceContext) registerRouterMap(server *rest.Server) {
	for _, item := range server.Routes() {
		if _, ok := sel.RouterMap[item.Method]; ok {
			sel.RouterMap[item.Method][item.Path] = item.Handler
		} else {
			sel.RouterMap[item.Method] = map[string]http.HandlerFunc{item.Path: item.Handler}
		}
	}
}

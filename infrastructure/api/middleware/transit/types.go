package transit

import (
	"net/http"

	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"
)

type (
	Middleware struct {
		IMiddlewareInterface
		routerMap     map[string]map[string]http.HandlerFunc
		encryptionKey string
		responder     responder.IResponder
		prevWidths    int
	}
)

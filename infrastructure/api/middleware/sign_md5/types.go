package sign_md5

import "github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"

type (
	Middleware struct {
		IMiddlewareInterface
		signName       string
		signKey        string
		filterMap      map[string]bool
		paramsSplit    string // default: &
		isSignKeySplit bool
		isParseArray   bool
		lastField      string
		responder      responder.IResponder
	}
)

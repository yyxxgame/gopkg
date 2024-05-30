package transit

import "github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"

type (
	Option func(*Middleware)
)

func WithResponder(responder responder.IResponder) Option {
	return func(m *Middleware) {
		m.responder = responder
	}
}

func WithPrevWidths(prevWidths int) Option {
	return func(m *Middleware) {
		m.prevWidths = prevWidths
	}
}

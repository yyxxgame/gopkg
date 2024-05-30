package sign_md5

import (
	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"
)

type (
	Option func(*Middleware)
)

func WithSignName(signName string) Option {
	return func(m *Middleware) {
		m.signName = signName
	}
}

func WithFilterFields(filterFields []string) Option {
	return func(m *Middleware) {
		for _, v := range filterFields {
			m.filterMap[v] = true
		}
	}
}

func WithParamsSplit(paramsSplit string) Option {
	return func(m *Middleware) {
		m.paramsSplit = paramsSplit
	}
}

func WithSignKeySplit(enable bool) Option {
	return func(m *Middleware) {
		m.isSignKeySplit = enable
	}
}

func WithParseArray(enable bool) Option {
	return func(m *Middleware) {
		m.isParseArray = enable
	}
}

func WithLastField(lastField string) Option {
	return func(m *Middleware) {
		m.lastField = lastField
	}
}

func WithResponder(responder responder.IResponder) Option {
	return func(m *Middleware) {
		m.responder = responder
	}
}

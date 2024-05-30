package getip

import (
	"context"
	"net/http"

	"github.com/thinkeridea/go-extend/exnet"
)

func New(ipKey string) IMiddlewareInterface {
	return &Middleware{
		ipKey: ipKey,
	}
}

func (m *Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set real ip to context with the specified name
		ip := exnet.ClientPublicIP(r)
		if ip == "" {
			ip = exnet.ClientIP(r)
		}
		ctx := context.WithValue(r.Context(), "__ipKey", m.ipKey) // nolint
		ctx = context.WithValue(ctx, m.ipKey, ip)                 // nolint
		next(w, r.WithContext(ctx))
	}
}

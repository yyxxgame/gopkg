package getip

import (
	"context"
	"net/http"

	"github.com/duke-git/lancet/v2/netutil"
)

func New(ipKey string) IMiddlewareInterface {
	return &Middleware{
		ipKey: ipKey,
	}
}

func (m *Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set real ip to context with the specified name
		ip := netutil.GetRequestPublicIp(r)
		if ip == "::1" {
			ip = "127.0.0.1"
		}
		ctx := context.WithValue(r.Context(), "__ipKey", m.ipKey) // nolint
		ctx = context.WithValue(ctx, m.ipKey, ip)
		next(w, r.WithContext(ctx))
	}
}

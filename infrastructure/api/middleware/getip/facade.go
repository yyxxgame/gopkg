package getip

import "net/http"

type (
	IMiddlewareInterface interface {
		Handle(next http.HandlerFunc) http.HandlerFunc
	}
)

package sign_md5

import "net/http"

type (
	IMiddlewareInterface interface {
		Handle(next http.HandlerFunc) http.HandlerFunc
	}
)

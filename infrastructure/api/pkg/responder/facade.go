package responder

import "net/http"

type (
	IResponder interface {
		GetEncryptionHeader() string
		Response(w http.ResponseWriter, r *http.Request, resp interface{}, err error)
	}
)

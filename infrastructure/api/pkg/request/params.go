package request

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
)

const (
	maxMemory = 32 << 20 // 32MB
)

func parseForm(r *http.Request) (map[string]interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return nil, err
		}
	}
	params := make(map[string]any, len(r.Form))
	for name := range r.Form {
		params[name] = r.Form.Get(name)
	}
	return params, nil
}

func GetAllParams(r *http.Request) map[string]any {
	params := make(map[string]any)

	cacheParams := r.Context().Value("__params").(string)
	if len(cacheParams) > 0 {
		_ = sonic.UnmarshalString(cacheParams, &params)
		return params
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	switch {
	case strings.Contains(contentType, "application/json"):
		bodyBytes, _ := io.ReadAll(r.Body)
		_ = sonic.Unmarshal(bodyBytes, &params)
	default:
		params, _ = parseForm(r)
	}

	return params
}

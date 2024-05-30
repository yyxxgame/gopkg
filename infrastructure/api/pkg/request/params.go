package request

import (
	"bytes"
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
		if err != http.ErrNotMultipart {
			return nil, err
		}
	}
	params := make(map[string]interface{}, len(r.Form))
	for name := range r.Form {
		params[name] = r.Form.Get(name)
	}
	return params, nil
}

func GetAllParams(r *http.Request) map[string]interface{} {
	params := make(map[string]interface{})
	cacheParams := r.Header.Get("__params")
	if len(cacheParams) > 0 {
		if err := sonic.UnmarshalString(cacheParams, &params); err != nil {
			return nil
		} else {
			return params
		}
	} else {
		formParams, _ := parseForm(r)
		for key, val := range formParams {
			params[key] = val
		}
		bodyBytes, _ := io.ReadAll(r.Body)
		contentType := r.Header["Content-Type"]
		if (contentType != nil && strings.Contains(contentType[0], "application/json")) || strings.Contains(string(bodyBytes), "{") {
			jsonParams := make(map[string]interface{})
			_ = sonic.Unmarshal(bodyBytes, &jsonParams)
			for key, val := range jsonParams {
				params[key] = val
			}
		}
		// Rewrite Request Body
		if paramsBytes, err := sonic.Marshal(params); err == nil {
			r.Header.Set("Content-Type", "application/json")
			buf := bytes.NewBuffer(paramsBytes)
			r.ContentLength = int64(buf.Len())
			r.Body = io.NopCloser(buf)
			// Cache Params To Header
			r.Header.Set("__params", string(paramsBytes))
		}
		return params
	}
}

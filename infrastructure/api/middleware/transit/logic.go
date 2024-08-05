package transit

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/yyxxgame/gopkg/cryptor/aes"
	"github.com/yyxxgame/gopkg/exception"
	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"
	"github.com/yyxxgame/gopkg/xtrace"
	gozerotrace "github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest/handler"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func New(routerMap map[string]map[string]http.HandlerFunc, encryptionKey string, options ...Option) IMiddlewareInterface {
	m := &Middleware{
		routerMap:     routerMap,
		encryptionKey: encryptionKey,
	}
	for _, option := range options {
		option(m)
	}
	if m.responder == nil {
		m.responder = responder.New()
	}
	return m
}

func (m *Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exception.Try(func() {
			headerIsDecrypt := m.responder.GetEncryptionHeader()
			if r.Header.Get(headerIsDecrypt) == "1" {
				r.Header.Set(headerIsDecrypt, "1")
			} else {
				r.Header.Set(headerIsDecrypt, "0")
			}
			bodyBytes, _ := io.ReadAll(r.Body)
			// Decrypt Body
			aesBodyString, err := aes.DecryptCbcPkcs7(hex.DecodeString, string(bodyBytes), m.encryptionKey)
			if err != nil {
				m.responder.Response(w, r, map[string]interface{}{"code": -3, "msg": "transit error"}, nil)
				return
			}
			jsonParams := make(map[string]interface{})
			_ = sonic.UnmarshalString(aesBodyString[m.prevWidths:], &jsonParams)
			var action string
			if maputil.HasKey(jsonParams, "action") {
				action = convertor.ToString(jsonParams["action"])
			}
			// Rewrite Request Body
			jsonParams["ch_conter"] = string(bodyBytes)
			if jsonParamsBytes, err := json.Marshal(jsonParams); err == nil {
				r.Header.Set("Content-Type", "application/json")
				r.Body = io.NopCloser(bytes.NewBuffer(jsonParamsBytes))
			}
			// Redirect next
			actionPath := fmt.Sprintf("/%s", action)
			nextCall := m.routerMap[r.Method][actionPath]
			if nextCall != nil && action != "" {
				_ = xtrace.WithTraceHook(r.Context(), otel.GetTracerProvider().Tracer(gozerotrace.TraceName), oteltrace.SpanKindInternal, action, func(ctx context.Context) error {
					handler.PrometheusHandler(actionPath, r.Method)(nextCall).ServeHTTP(w, r)
					return nil
				})
			} else {
				next(w, r)
			}
		}).Catch(func(e exception.Exception) {
			next(w, r)
		}).Finally(func() {})
	}
}

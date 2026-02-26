package responder

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/yyxxgame/gopkg/cryptor/aes"
	"github.com/yyxxgame/gopkg/infrastructure/api/pkg"
	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/request"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
)

func New(options ...Option) *Responder {
	m := &Responder{}
	for _, opt := range options {
		opt(m)
	}
	return m
}

func (m *Responder) GetEncryptionHeader() string {
	return m.encryptionHeader
}

func (m *Responder) logDetails(r *http.Request, resp any, err error) {
	reqParams := request.GetAllParams(r)
	ip := pkg.GetIpFromContext(r.Context())
	accessLog := "path:(%s)ip(%s)param:(%v)return:(%#v)"
	accessLog = fmt.Sprintf(accessLog, r.URL.Path, ip, reqParams, resp)
	logx.WithContext(r.Context()).Info(accessLog)
	xtrace.AddEvent(
		r.Context(), r.URL.Path,
		attribute.String("req.gmip", ip),
		attribute.String("req.body.raw", convertor.ToString(reqParams)),
		attribute.String("resp.body.raw", convertor.ToString(resp)),
	)
	if err != nil {
		logx.WithContext(r.Context()).Errorf(accessLog + fmt.Sprintf("err:(%s)", err))
		xtrace.AddEvent(r.Context(), r.RequestURI, attribute.String("err", err.Error()))
	}
}

func (m *Responder) Response(w http.ResponseWriter, r *http.Request, resp any, err error) {
	if m.encryptionHeader != "" && r.Header.Get(m.encryptionHeader) == "0" {
		bf := bytes.NewBuffer([]byte{})
		jsonEncoder := json.NewEncoder(bf)
		jsonEncoder.SetEscapeHTML(false)
		err = jsonEncoder.Encode(resp)
		if err != nil && resp == nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			reqString := string(bf.Bytes()[:bf.Len()-1]) // Remove the '\n' at the end
			bs, err := aes.EncryptCbcPkcs7(hex.EncodeToString, reqString, m.encryptionKey)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			} else {
				if n, err := w.Write([]byte(bs)); err != nil {
					if !errors.Is(err, http.ErrHandlerTimeout) {
						logx.WithContext(r.Context()).Errorf("write response failed, error: %s", err)
					}
				} else if n < len(bs) {
					logx.WithContext(r.Context()).Errorf("actual bytes: %d, written bytes: %d", len(bs), n)
				}
			}
		}
	} else {
		if err != nil && resp == nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			switch v := resp.(type) {
			case string:
				_, err = w.Write([]byte(v))
			case []byte:
				_, err = w.Write(v)
			default:
				httpx.OkJsonCtx(r.Context(), w, v)
			}
		}
	}
	// Log Details
	if m.accessLogDetails {
		m.logDetails(r, resp, err)
	}
}

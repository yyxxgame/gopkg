package sign_md5

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/request"
	"github.com/yyxxgame/gopkg/infrastructure/api/pkg/responder"
	"github.com/zeromicro/go-zero/core/hash"
)

func New(signKey string, options ...Option) IMiddlewareInterface {
	m := &Middleware{
		signKey:   signKey,
		filterMap: map[string]bool{},
	}
	for _, option := range options {
		option(m)
	}
	if len(m.signName) == 0 {
		m.signName = "sign"
	}
	m.filterMap[m.signName] = true
	if len(m.paramsSplit) == 0 {
		m.paramsSplit = "&"
	}
	if len(m.lastField) > 0 {
		m.filterMap[m.lastField] = true
	}
	if m.responder == nil {
		m.responder = responder.New()
	}
	return m
}

func (m *Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := request.GetAllParams(r)
		if !m.checkSign(req) {
			m.responder.Response(w, r, map[string]interface{}{"code": -3, "msg": "sign error"}, nil)
			return
		}
		next(w, r)
	}
}

func (m *Middleware) checkSign(req map[string]interface{}) bool {
	signStr := m.makeSign(req)
	return req[m.signName] == signStr
}

func (m *Middleware) makeSign(values map[string]interface{}) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	signParams := make([]string, 0, len(values))
	for _, key := range keys {
		if _, ok := m.filterMap[key]; ok {
			continue
		}
		if m.isParseArray {
			v := values[key]
			switch v := v.(type) {
			case []interface{}:
				if len(v) == 1 {
					values[key] = v[0]
				} else {
					raw, _ := sonic.Marshal(v)
					values[key] = string(raw)
				}
			}
		}
		signParams = append(signParams, fmt.Sprintf("%s=%v", key, convertor.ToString(values[key])))
	}
	if m.lastField != "" {
		signParams = append(signParams, fmt.Sprintf("%s=%v", m.lastField, convertor.ToString(values[m.lastField])))
	}
	signKeySplit := ""
	if m.isSignKeySplit {
		signKeySplit = m.paramsSplit
	}
	signStr := fmt.Sprintf("%s%s%s", strings.Join(signParams, m.paramsSplit), signKeySplit, m.signKey)
	sign := hash.Md5Hex([]byte(signStr))
	return sign
}

package middleware

import (
	"net/http"

	"github.com/yyxxgame/gopkg/infrastructure/api/middleware/getip"
	signMd5 "github.com/yyxxgame/gopkg/infrastructure/api/middleware/sign_md5"
	"github.com/yyxxgame/gopkg/infrastructure/api/middleware/transit"
	yyxxSvc "github.com/yyxxgame/gopkg/infrastructure/api/svc"
	"github.com/zeromicro/go-zero/rest"
)

type (
	Builder struct {
		*yyxxSvc.ServiceContext
		FilterFields []string
	}
)

func NewBuilder(m *yyxxSvc.ServiceContext) *Builder {
	builder := &Builder{
		ServiceContext: m,
		FilterFields:   []string{"gmip", "cp_platform", "ch_conter", "opts", "is_whdon"},
	}
	return builder
}

func (b *Builder) Build() {
	b.BuildIp()

	if len(b.Config.EncryptionKey) > 0 {
		b.BuildTransit()
	}

	if len(b.Config.ApiKey) > 0 {
		b.BuildSignMd5(
			&b.AuthMiddleware, b.Config.ApiKey,
			signMd5.WithLastField("time"),
		)
	}
}

func (b *Builder) BuildIp() {
	b.GetIpMiddleware = getip.New("gmip").Handle
}

func (b *Builder) BuildTransit() {
	b.RouterMap = make(map[string]map[string]http.HandlerFunc)
	b.TransitMiddleware = transit.New(
		b.RouterMap, b.Config.EncryptionKey,
		transit.WithPrevWidths(5),
		transit.WithResponder(b.Responder),
	).Handle
}

func (b *Builder) BuildSignMd5(src *rest.Middleware, signKey string, options ...signMd5.Option) {
	options = append(
		options,
		signMd5.WithFilterFields(b.FilterFields),
		signMd5.WithResponder(b.Responder),
	)
	*src = signMd5.New(signKey, options...).Handle
}

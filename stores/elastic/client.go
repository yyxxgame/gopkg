package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	Option func(c *client)
	client struct {
		username   string
		password   string
		enableAuth bool
		enableGzip bool
		*elastic.Client
		tracer oteltrace.Tracer
	}

	EsSearchChain func(srv *elastic.SearchService) *elastic.SearchService

	IEsClient interface {
		Query(chain EsSearchChain) *elastic.SearchResult
		QueryCtx(ctx context.Context, chain EsSearchChain) *elastic.SearchResult
		Insert(index string, data interface{})
		InsertCtx(ctx context.Context, index string, data interface{})
	}
)

func MustNew(endpoints []string, opts ...Option) IEsClient {
	impl := &client{}
	for _, opt := range opts {
		opt(impl)
	}
	if cli, err := elastic.NewSimpleClient(
		elastic.SetURL(endpoints...),
		elastic.SetGzip(impl.enableGzip),
		elastic.SetBasicAuth(impl.username, impl.password),
	); err != nil {
		logx.WithContext(context.Background()).Error(err.Error())
		panic(err.Error())
	} else {
		impl.Client = cli
	}
	return impl
}

func WithAuth(username, password string) Option {
	return func(c *client) {
		c.enableAuth = true
		c.username = username
		c.password = password
	}
}

func WithGzip() Option {
	return func(c *client) {
		c.enableGzip = true
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *client) {
		c.tracer = tracer
	}
}

func (c *client) Query(chain EsSearchChain) *elastic.SearchResult {
	return c.QueryCtx(context.Background(), chain)
}

func (c *client) QueryCtx(ctx context.Context, chain EsSearchChain) *elastic.SearchResult {
	var result *elastic.SearchResult
	if c.tracer != nil {
		xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "elastic.QueryCtx", func(ctx context.Context) error {
			if ret, err := chain(c.Search()).Pretty(true).Do(ctx); err != nil {
				logx.WithContext(ctx).Errorf("query data on error: %v", err)
				return err
			} else {
				result = ret
			}
			return nil
		})
	} else {
		if ret, err := chain(c.Search()).Pretty(true).Do(ctx); err != nil {
			logx.WithContext(ctx).Errorf("query data on error: %v", err)
		} else {
			result = ret
		}
	}
	return result
}

func (c *client) Insert(index string, data interface{}) {
	c.InsertCtx(context.Background(), index, data)
}

func (c *client) InsertCtx(ctx context.Context, index string, data interface{}) {
	if c.tracer != nil {
		xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "elastic.InsertCtx", func(ctx context.Context) error {
			if result, err := c.Index().Index(index).Refresh("false").BodyJson(data).Do(ctx); err != nil {
				logx.WithContext(ctx).Errorf("insert data on error: %v", err)
				return err
			} else {
				logx.WithContext(ctx).Infof("insert data on success: %+v", *result)
			}
			return nil
		})
	} else {
		if result, err := c.Index().Index(index).Refresh("false").BodyJson(data).Do(ctx); err != nil {
			logx.WithContext(ctx).Errorf("insert data on error: %v", err)
		} else {
			logx.WithContext(ctx).Infof("insert data on success: %+v", *result)
		}
	}
}

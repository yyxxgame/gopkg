package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/yyxxgame/gopkg/xtrace"
	"github.com/zeromicro/go-zero/core/logx"
)

type (
	Option func(c *client)
	client struct {
		username   string
		password   string
		enableAuth bool
		enableGzip bool
		instance   *elastic.Client
	}

	IEsClient interface {
		Query(handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult
		QueryCtx(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult
		query(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult
		Insert(index string, data interface{})
		InsertCtx(ctx context.Context, index string, data interface{})
		insert(ctx context.Context, index string, data interface{})
	}
)

func New(endpoints []string, opts ...Option) IEsClient {
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
		impl.instance = cli
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

func WithGzip(enable bool) Option {
	return func(c *client) {
		c.enableGzip = enable
	}
}

func (impl *client) Query(handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult {
	return impl.QueryCtx(context.Background(), handle)
}

func (impl *client) QueryCtx(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult {
	var result *elastic.SearchResult
	xtrace.StartFuncSpan(ctx, "QueryDataFromEs", func(ctx context.Context) {
		result = impl.query(ctx, handle)
	})
	return result
}

func (impl *client) query(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult {
	impl.instance.IngestGetPipeline()
	if result, err := handle(impl.instance.Search()).Pretty(true).Do(ctx); err != nil {
		logx.WithContext(ctx).Errorf("query data on error: %v", err)
		return nil
	} else {
		return result
	}
}

func (impl *client) Insert(index string, data interface{}) {
	impl.InsertCtx(context.Background(), index, data)
}

func (impl *client) InsertCtx(ctx context.Context, index string, data interface{}) {
	xtrace.StartFuncSpan(ctx, "InsertDataToEs", func(ctx context.Context) {
		impl.insert(ctx, index, data)
	})
}

func (impl *client) insert(ctx context.Context, index string, data interface{}) {
	if result, err := impl.instance.Index().Index(index).Refresh("false").BodyJson(data).Do(ctx); err != nil {
		logx.WithContext(ctx).Errorf("insert data on error: %v", err)
	} else {
		logx.WithContext(ctx).Infof("insert data on success: %+v", *result)
	}
}

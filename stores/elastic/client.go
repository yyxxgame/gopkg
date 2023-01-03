package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/yyxxgame/gopkg/trace"
	"github.com/zeromicro/go-zero/core/logx"
)

type IEsClient interface {
	Query(handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult
	QueryCtx(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult
	query(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult
	Insert(index string, data interface{})
	InsertCtx(ctx context.Context, index string, data interface{})
	insert(ctx context.Context, index string, data interface{})
}

type (
	Option     func(impl *clientImpl)
	clientImpl struct {
		username   string
		password   string
		enableAuth bool
		enableGzip bool
		client     *elastic.Client
	}
)

func NewEsClient(endpoints []string, opts ...Option) IEsClient {
	impl := &clientImpl{}
	for _, opt := range opts {
		opt(impl)
	}
	cli, err := elastic.NewSimpleClient(
		elastic.SetURL(endpoints...),
		elastic.SetGzip(impl.enableGzip),
		elastic.SetBasicAuth(impl.username, impl.password),
	)

	if err != nil {
		logx.WithContext(context.Background()).Error(err.Error())
		panic(err.Error())
	}
	impl.client = cli
	return impl
}

func WithAuth(username, password string) Option {
	return func(es *clientImpl) {
		es.username = username
		es.password = password
	}
}

func WithGzip(enable bool) Option {
	return func(es *clientImpl) {
		es.enableGzip = enable
	}
}

func (impl *clientImpl) Query(handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult {
	return impl.QueryCtx(context.Background(), handle)
}

func (impl *clientImpl) QueryCtx(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult {
	var result *elastic.SearchResult
	trace.StartFuncSpan(ctx, "QueryDataFromEs", func(ctx context.Context) {
		result = impl.query(ctx, handle)
	})
	return result
}

func (impl *clientImpl) query(ctx context.Context, handle func(srv *elastic.SearchService) *elastic.SearchService) *elastic.SearchResult {
	impl.client.IngestGetPipeline()
	if result, err := handle(impl.client.Search()).Pretty(true).Do(ctx); err != nil {
		logx.WithContext(ctx).Errorf("query data on error: %v", err)
		return nil
	} else {
		return result
	}
}

func (impl *clientImpl) Insert(index string, data interface{}) {
	impl.InsertCtx(context.Background(), index, data)
}

func (impl *clientImpl) InsertCtx(ctx context.Context, index string, data interface{}) {
	trace.StartFuncSpan(ctx, "InsertDataToEs", func(ctx context.Context) {
		impl.insert(ctx, index, data)
	})
}

func (impl *clientImpl) insert(ctx context.Context, index string, data interface{}) {
	if result, err := impl.client.Index().Index(index).Refresh("false").BodyJson(data).Do(ctx); err != nil {
		logx.WithContext(ctx).Errorf("insert data on error: %v", err)
	} else {
		logx.WithContext(ctx).Infof("insert data on success: %+v", *result)
	}
}

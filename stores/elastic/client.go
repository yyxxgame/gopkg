package elastic

import (
	"context"
	v7elastic "github.com/olivere/elastic/v7"
	"github.com/yyxxgame/gopkg/xtrace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	QueryChain  func(srv *v7elastic.SearchService) *v7elastic.SearchService
	InsertChain func(srv *v7elastic.IndexService) *v7elastic.IndexService
	UpsertChain func(srv *v7elastic.UpdateService) *v7elastic.UpdateService

	IEsClient interface {
		Query(chain QueryChain) (*v7elastic.SearchResult, error)
		QueryCtx(ctx context.Context, chain QueryChain) (*v7elastic.SearchResult, error)
		doQuery(ctx context.Context, chain QueryChain) (*v7elastic.SearchResult, error)
		Insert(chain InsertChain) (*v7elastic.IndexResponse, error)
		InsertCtx(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error)
		doInsert(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error)
		Upsert(chain UpsertChain) (*v7elastic.UpdateResponse, error)
		UpsertCtx(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error)
		doUpsert(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error)
	}

	Option func(c *esClient)

	esClient struct {
		username   string
		password   string
		enableGzip bool
		tracer     oteltrace.Tracer

		*v7elastic.Client
	}
)

func MustNew(endpoints []string, opts ...Option) IEsClient {
	c := &esClient{}
	for _, opt := range opts {
		opt(c)
	}
	if client, err := v7elastic.NewSimpleClient(
		v7elastic.SetURL(endpoints...),
		v7elastic.SetGzip(c.enableGzip),
		v7elastic.SetBasicAuth(c.username, c.password),
	); err != nil {
		panic(err.Error())
	} else {
		c.Client = client
	}
	return c
}

func (c *esClient) Query(chain QueryChain) (*v7elastic.SearchResult, error) {
	return c.QueryCtx(context.Background(), chain)
}

func (c *esClient) QueryCtx(ctx context.Context, chain QueryChain) (*v7elastic.SearchResult, error) {
	if c.tracer == nil {
		return c.doQuery(ctx, chain)
	}

	var resp *v7elastic.SearchResult
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "v7elastic.Query", func(ctx context.Context) error {
		if ret, err := c.doQuery(ctx, chain); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) doQuery(ctx context.Context, chain QueryChain) (*v7elastic.SearchResult, error) {
	if resp, err := chain(v7elastic.NewSearchService(c.Client)).Pretty(true).Do(ctx); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *esClient) Insert(chain InsertChain) (*v7elastic.IndexResponse, error) {
	return c.InsertCtx(context.Background(), chain)
}

func (c *esClient) InsertCtx(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error) {
	if c.tracer == nil {
		return c.doInsert(ctx, chain)
	}

	var resp *v7elastic.IndexResponse
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "v7elastic.Insert", func(ctx context.Context) error {
		if ret, err := c.doInsert(ctx, chain); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) doInsert(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error) {
	if resp, err := chain(v7elastic.NewIndexService(c.Client)).Pretty(true).Refresh("true").Do(ctx); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (c *esClient) Upsert(chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	return c.UpsertCtx(context.Background(), chain)
}

func (c *esClient) UpsertCtx(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	if c.tracer == nil {
		return c.doUpsert(ctx, chain)
	}

	var resp *v7elastic.UpdateResponse
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "v7elastic.Insert", func(ctx context.Context) error {
		if ret, err := c.doUpsert(ctx, chain); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) doUpsert(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	if resp, err := chain(v7elastic.NewUpdateService(c.Client)).DocAsUpsert(true).Pretty(true).Refresh("true").Do(ctx); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func WithAuth(username, password string) Option {
	return func(c *esClient) {
		c.username = username
		c.password = password
	}
}

func WithGzip() Option {
	return func(c *esClient) {
		c.enableGzip = true
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(c *esClient) {
		c.tracer = tracer
	}
}

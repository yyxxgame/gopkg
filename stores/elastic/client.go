package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/yyxxgame/gopkg/xtrace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	QueryChain  func(srv *elastic.SearchService) *elastic.SearchService
	InsertChain func(srv *elastic.IndexService) *elastic.IndexService

	IEsClient interface {
		Query(chain QueryChain) (*elastic.SearchResult, error)
		QueryCtx(ctx context.Context, chain QueryChain) (*elastic.SearchResult, error)
		doQuery(ctx context.Context, chain QueryChain) (*elastic.SearchResult, error)
		Insert(chain InsertChain) (*elastic.IndexResponse, error)
		InsertCtx(ctx context.Context, chain InsertChain) (*elastic.IndexResponse, error)
		doInsert(ctx context.Context, chain InsertChain) (*elastic.IndexResponse, error)
	}

	Option func(c *esClient)

	esClient struct {
		username   string
		password   string
		enableGzip bool
		tracer     oteltrace.Tracer

		*elastic.Client
	}
)

func MustNew(endpoints []string, opts ...Option) IEsClient {
	c := &esClient{}
	for _, opt := range opts {
		opt(c)
	}
	if client, err := elastic.NewSimpleClient(
		elastic.SetURL(endpoints...),
		elastic.SetGzip(c.enableGzip),
		elastic.SetBasicAuth(c.username, c.password),
	); err != nil {
		panic(err.Error())
	} else {
		c.Client = client
	}
	return c
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

func (c *esClient) Query(chain QueryChain) (*elastic.SearchResult, error) {
	return c.QueryCtx(context.Background(), chain)
}

func (c *esClient) QueryCtx(ctx context.Context, chain QueryChain) (*elastic.SearchResult, error) {
	if c.tracer != nil {
		var resp *elastic.SearchResult
		err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "elastic.Query", func(ctx context.Context) error {
			if ret, err := c.doQuery(ctx, chain); err != nil {
				return err
			} else {
				resp = ret
			}
			return nil
		})
		return resp, err
	} else {
		return c.doQuery(ctx, chain)
	}
}

func (c *esClient) doQuery(ctx context.Context, chain QueryChain) (*elastic.SearchResult, error) {
	if resp, err := chain(c.Search()).Pretty(true).Do(ctx); err != nil {
		return nil, err
	} else {
		return resp, err
	}
}

func (c *esClient) Insert(chain InsertChain) (*elastic.IndexResponse, error) {
	return c.InsertCtx(context.Background(), chain)
}

func (c *esClient) InsertCtx(ctx context.Context, chain InsertChain) (*elastic.IndexResponse, error) {
	if c.tracer != nil {
		var resp *elastic.IndexResponse
		err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindInternal, "elastic.Insert", func(ctx context.Context) error {
			if ret, err := c.doInsert(ctx, chain); err != nil {
				return err
			} else {
				resp = ret
			}
			return nil
		})
		return resp, err
	} else {
		return c.doInsert(ctx, chain)
	}
}

func (c *esClient) doInsert(ctx context.Context, chain InsertChain) (*elastic.IndexResponse, error) {
	if resp, err := chain(c.Index()).Refresh("false").Do(ctx); err != nil {
		return nil, err
	} else {
		return resp, err
	}
}

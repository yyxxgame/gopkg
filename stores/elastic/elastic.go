package elastic

import (
	"context"

	v7elastic "github.com/olivere/elastic/v7"
	"github.com/yyxxgame/gopkg/xtrace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type (
	QueryChain    func(srv *v7elastic.SearchService) *v7elastic.SearchService
	InsertChain   func(srv *v7elastic.IndexService) *v7elastic.IndexService
	UpsertChain   func(srv *v7elastic.UpdateService) *v7elastic.UpdateService
	BulkExecChain func(srv *v7elastic.BulkService) *v7elastic.BulkService
	AnalyzeChain  func(srv *v7elastic.IndicesAnalyzeService) *v7elastic.IndicesAnalyzeService

	IEsClient interface {
		Query(chain QueryChain) (*v7elastic.SearchResult, error)
		QueryCtx(ctx context.Context, chain QueryChain) (*v7elastic.SearchResult, error)
		Insert(chain InsertChain) (*v7elastic.IndexResponse, error)
		InsertCtx(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error)
		Upsert(chain UpsertChain) (*v7elastic.UpdateResponse, error)
		UpsertCtx(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error)
		BulkExec(chain BulkExecChain) (*v7elastic.BulkResponse, error)
		BulkExecCtx(ctx context.Context, chain BulkExecChain) (*v7elastic.BulkResponse, error)
		Analyze(chain AnalyzeChain) (*v7elastic.IndicesAnalyzeResponse, error)
		AnalyzeCtx(ctx context.Context, chain AnalyzeChain) (*v7elastic.IndicesAnalyzeResponse, error)
	}

	esClient struct {
		username          string
		password          string
		enableGzip        bool
		enableHealthCheck bool
		enableSniffer     bool
		tracer            oteltrace.Tracer

		*v7elastic.Client
	}
)

func MustNew(endpoints []string, opts ...Option) IEsClient {
	c := &esClient{}
	for _, opt := range opts {
		opt(c)
	}
	if client, err := v7elastic.NewClient(
		v7elastic.SetURL(endpoints...),
		v7elastic.SetGzip(c.enableGzip),
		v7elastic.SetBasicAuth(c.username, c.password),
		v7elastic.SetHealthcheck(c.enableHealthCheck),
		v7elastic.SetSniff(c.enableSniffer),
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
		return chain(v7elastic.NewSearchService(c.Client)).Pretty(true).Do(ctx)
	}

	var resp *v7elastic.SearchResult
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindClient, "v7elastic.Query", func(ctx context.Context) error {
		if ret, err := chain(v7elastic.NewSearchService(c.Client)).Pretty(true).Do(ctx); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) Insert(chain InsertChain) (*v7elastic.IndexResponse, error) {
	return c.InsertCtx(context.Background(), chain)
}

func (c *esClient) InsertCtx(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error) {
	if c.tracer == nil {
		return chain(v7elastic.NewIndexService(c.Client)).Pretty(true).Refresh("true").Do(ctx)
	}

	var resp *v7elastic.IndexResponse
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindClient, "v7elastic.Insert", func(ctx context.Context) error {
		if ret, err := chain(v7elastic.NewIndexService(c.Client)).Pretty(true).Refresh("true").Do(ctx); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) Upsert(chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	return c.UpsertCtx(context.Background(), chain)
}

func (c *esClient) UpsertCtx(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	if c.tracer == nil {
		return chain(v7elastic.NewUpdateService(c.Client)).DocAsUpsert(true).Pretty(true).Refresh("true").Do(ctx)
	}

	var resp *v7elastic.UpdateResponse
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindClient, "v7elastic.Insert", func(ctx context.Context) error {
		if ret, err := chain(v7elastic.NewUpdateService(c.Client)).DocAsUpsert(true).Pretty(true).Refresh("true").Do(ctx); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) BulkExec(chain BulkExecChain) (*v7elastic.BulkResponse, error) {
	return c.BulkExecCtx(context.Background(), chain)
}

func (c *esClient) BulkExecCtx(ctx context.Context, chain BulkExecChain) (*v7elastic.BulkResponse, error) {
	if c.tracer == nil {
		return chain(v7elastic.NewBulkService(c.Client)).Pretty(true).Refresh("true").Do(ctx)
	}
	var resp *v7elastic.BulkResponse
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindClient, "v7elastic.BulkExec", func(ctx context.Context) error {
		if ret, err := chain(v7elastic.NewBulkService(c.Client)).Pretty(true).Refresh("true").Do(ctx); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

func (c *esClient) Analyze(chain AnalyzeChain) (*v7elastic.IndicesAnalyzeResponse, error) {
	return c.AnalyzeCtx(context.Background(), chain)
}

func (c *esClient) AnalyzeCtx(ctx context.Context, chain AnalyzeChain) (*v7elastic.IndicesAnalyzeResponse, error) {
	if c.tracer == nil {
		return chain(v7elastic.NewIndicesAnalyzeService(c.Client)).Pretty(true).Do(ctx)
	}

	var resp *v7elastic.IndicesAnalyzeResponse
	err := xtrace.WithTraceHook(ctx, c.tracer, oteltrace.SpanKindClient, "v7elastic.Analyze", func(ctx context.Context) error {
		if ret, err := chain(v7elastic.NewIndicesAnalyzeService(c.Client)).Pretty(true).Do(ctx); err != nil {
			return err
		} else {
			resp = ret
		}
		return nil
	})
	return resp, err
}

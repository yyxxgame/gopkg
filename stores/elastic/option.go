//@File     option.go
//@Time     2024/03/11
//@Author   #Suyghur,

package elastic

import oteltrace "go.opentelemetry.io/otel/trace"

type Option func(c *esClient)

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

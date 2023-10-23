//@File     option.go
//@Time     2023/10/23
//@Author   #Suyghur,

package gormc

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

const (
	defaultExpiry         = time.Hour * 24
	defaultNotFoundExpiry = time.Minute
)

type (
	// An Options is used to store the cache options.
	Options struct {
		//Expiry         time.Duration
		//NotFoundExpiry time.Duration
		cache.Options
		tracer       oteltrace.Tracer
		enableMetric bool
	}

	// Option defines the method to customize an Options.
	Option func(o *Options)
)

func newOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}

	if o.Expiry <= 0 {
		o.Expiry = defaultExpiry
	}
	if o.NotFoundExpiry <= 0 {
		o.NotFoundExpiry = defaultNotFoundExpiry
	}

	return o
}

// WithExpiry returns a func to customize an Options with given expiry.
func WithExpiry(expiry time.Duration) Option {
	return func(o *Options) {
		o.Expiry = expiry
	}
}

// WithNotFoundExpiry returns a func to customize an Options with given not found expiry.
func WithNotFoundExpiry(expiry time.Duration) Option {
	return func(o *Options) {
		o.NotFoundExpiry = expiry
	}
}

func WithTracer(tracer oteltrace.Tracer) Option {
	return func(o *Options) {
		o.tracer = tracer
	}
}

func WithMetric() Option {
	return func(o *Options) {
		o.enableMetric = true
	}
}

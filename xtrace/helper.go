//@File     : helper.go
//@Time     : 2023/2/21
//@Auther   : Kaishin

package xtrace

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func AddTags(ctx context.Context, kv ...attribute.KeyValue) {
	/*
		add tags by ctx
	*/
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
}

func AddEvent(ctx context.Context, name string, kv ...attribute.KeyValue) {
	/*
		add evnet by ctx
	*/
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(kv...))
}

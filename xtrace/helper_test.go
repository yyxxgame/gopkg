//@File     : helper_test.go
//@Time     : 2023/11/20
//@Auther   : Kaishin

package xtrace

import (
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"testing"
)

func TestMakeHeadContext(t *testing.T) {
	// 创建一个新的 TracerProvider
	tp := sdktrace.NewTracerProvider()

	// 设置全局的 TracerProvider
	otel.SetTracerProvider(tp)

	ctx, span := MakeHeaderContext("trace-test")
	defer span.End()
	t.Log(GetTraceId(ctx))
}

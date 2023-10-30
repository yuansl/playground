package trace

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/yuansl/playground/util"
)

type TracerProvider interface {
	trace.TracerProvider
	Shutdown(context.Context) error
}

var lazyCreateTracer func() TracerProvider

func init() {
	lazyCreateTracer = sync.OnceValue(func() TracerProvider {
		spanExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			util.Fatal(err)
		}
		_ = sdktrace.WithBatcher(spanExporter) // TODO
		tracerProvider := sdktrace.NewTracerProvider()

		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.TraceContext{})

		return tracerProvider
	})
}

func GetTracerProvider() TracerProvider {
	return lazyCreateTracer()
}

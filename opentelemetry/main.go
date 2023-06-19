package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

var _ propagation.TextMapPropagator

// Package-level tracer.
// This should be configured in your code setup instead of here.
var tracer trace.Tracer

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}

func init() {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		fatal("stdouttrace.New error: %v\n", err)
	}
	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample())))

	tracer = otel.Tracer("github.com/yuansl/playground/opentelemetry")
}

func stuff(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "stuff")
	defer span.End()

	span.AddEvent("calling do_something...", trace.WithStackTrace(true))

	startAt := time.Now()
	if err := do_something(ctx); err != nil {
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(400)...)
	} else {
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(200)...)
	}
	span.SetAttributes(attribute.Int("sleep.duration", int(time.Since(startAt))))
}

// do_something mocks work that your application does.
func do_something(ctx context.Context, hooks ...func(context.Context)) error {
	for _, hook := range hooks {
		hook(ctx)
	}

	httpctx := httpFromContext(ctx)

	fmt.Fprintf(httpctx.w, "Hello, World! I am instrumented automatically!\n")

	time.Sleep(500 * time.Millisecond)

	return nil
}

type httpContext struct {
	w   http.ResponseWriter
	req *http.Request
}

const httpContextKey = 0

func contextWithHttp(ctx context.Context, h *httpContext) context.Context {
	return context.WithValue(ctx, httpContextKey, h)
}

func httpFromContext(ctx context.Context) *httpContext {
	if c, ok := ctx.Value(httpContextKey).(*httpContext); ok {
		return c
	}
	panic("BUG: can't extract httpContext from the ctx")
}

func main() {
	// Wrap your httpHandler function.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := contextWithHttp(r.Context(), &httpContext{w: w, req: r})

		stuff(ctx)
	})

	wrappedHandler := otelhttp.NewHandler(handler, "http.path: /hello-instrumented",
		otelhttp.WithPropagators(propagation.TraceContext{}))

	http.Handle("/hello-instrumented", wrappedHandler)

	// And start the HTTP serve.
	log.Fatal(http.ListenAndServe(":3030", nil))
}

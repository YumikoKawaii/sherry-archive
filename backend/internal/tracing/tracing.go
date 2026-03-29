package tracing

import (
	"context"
	"fmt"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/yumikokawaii/sherry-archive/internal/config"
)

// Init sets up the OpenTelemetry tracer provider with a Jaeger OTLP exporter.
// Returns a shutdown function that flushes pending spans and a SQL driver name
// to pass to postgres.ConnectWithDriver.
//
// When disabled, returns a no-op shutdown and the plain "postgres" driver name
// so the rest of serve.go works unchanged.
func Init(cfg *config.TracingConfig) (shutdown func(), driverName string, err error) {
	noop := func() {}

	if !cfg.Enabled {
		return noop, "postgres", nil
	}

	// Register an instrumented wrapper around lib/pq under a new driver name.
	driverName, err = otelsql.Register("postgres",
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithSpanOptions(otelsql.SpanOptions{DisableErrSkip: true}),
	)
	if err != nil {
		return noop, "postgres", fmt.Errorf("otelsql register: %w", err)
	}

	// OTLP HTTP exporter — points at Jaeger (or OTel Collector).
	// cfg.Endpoint is host:port, e.g. "jaeger-host:4318".
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return noop, driverName, fmt.Errorf("otlp exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("sherry-archive"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() { _ = tp.Shutdown(context.Background()) }, driverName, nil
}

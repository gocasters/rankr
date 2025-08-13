package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// otelTracerProvider handles tracing functionality
type otelTracerProvider struct {
	provider     *sdktrace.TracerProvider
	isConfigured bool
}

func newOtelTracerProvider(exporterFactory ExporterFactory, resourceBuilder ResourceBuilder, serviceName string) (*otelTracerProvider, error) {
	ctx := context.Background()

	exp, err := exporterFactory.CreateTraceExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resourceBuilder.BuildResource(serviceName)),
	)

	otel.SetTracerProvider(tp)

	return &otelTracerProvider{
		provider:     tp,
		isConfigured: true,
	}, nil
}

func (t *otelTracerProvider) NewTracer(name string, options ...trace.TracerOption) trace.Tracer {
	if !t.isConfigured {
		panic("tracer provider not configured")
	}
	return t.provider.Tracer(name, options...)
}

func (t *otelTracerProvider) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

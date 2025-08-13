package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

// otelMetricProvider handles metrics functionality
type otelMetricProvider struct {
	provider     *sdkMetric.MeterProvider
	isConfigured bool
}

func newOtelMetricProvider(exporterFactory ExporterFactory, resourceBuilder ResourceBuilder, serviceName string) (*otelMetricProvider, error) {
	ctx := context.Background()

	exp, err := exporterFactory.CreateMetricExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	mp := sdkMetric.NewMeterProvider(
		sdkMetric.WithReader(sdkMetric.NewPeriodicReader(exp)),
		sdkMetric.WithResource(resourceBuilder.BuildResource(serviceName)),
	)

	otel.SetMeterProvider(mp)

	return &otelMetricProvider{
		provider:     mp,
		isConfigured: true,
	}, nil
}

func (m *otelMetricProvider) NewMeter(name string, options ...metric.MeterOption) metric.Meter {
	if !m.isConfigured {
		panic("metric provider not configured")
	}
	return m.provider.Meter(name, options...)
}

func (m *otelMetricProvider) Shutdown(ctx context.Context) error {
	if m.provider != nil {
		return m.provider.Shutdown(ctx)
	}
	return nil
}

package otel

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// TracerProvider abstracts the tracing functionality
type TracerProvider interface {
	NewTracer(name string, options ...trace.TracerOption) (trace.Tracer, error)
	Shutdown(ctx context.Context) error
}

// MetricProvider abstracts the metrics functionality
type MetricProvider interface {
	NewMeter(name string, options ...metric.MeterOption) metric.Meter
	Shutdown(ctx context.Context) error
}

// PropagationHandler abstracts context propagation
type PropagationHandler interface {
	GetCarrierFromContext(ctx context.Context) map[string]string
	GetContextFromCarrier(carrier map[string]string) context.Context
}

// ConnectionManager abstracts connection management
type ConnectionManager interface {
	GetConnection() *grpc.ClientConn
	Close() error
}

// ExporterFactory creates exporters based on configuration
type ExporterFactory interface {
	CreateTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error)
	CreateMetricExporter(ctx context.Context) (sdkMetric.Exporter, error)
}

// ResourceBuilder creates OpenTelemetry resources
type ResourceBuilder interface {
	BuildResource(serviceName string) *resource.Resource
}

// OtelAdapter combines all telemetry functionality
type OtelAdapter interface {
	TracerProvider
	MetricProvider
	PropagationHandler
	IsConfigured() bool
	AddFloat64Counter(ctx context.Context, meter metric.Meter, name, desc string, cv float64, options ...metric.Float64CounterOption) error
	Shutdown(ctx context.Context) error
}

package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// otelExporterFactory creates exporters based on configuration
type otelExporterFactory struct {
	exporter Exporter
	connMgr  ConnectionManager
}

func newOtelExporterFactory(exporter Exporter, connMgr ConnectionManager) ExporterFactory {
	return &otelExporterFactory{
		exporter: exporter,
		connMgr:  connMgr,
	}
}

func (f *otelExporterFactory) CreateTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	switch f.exporter {
	case ExporterGrpc:
		return otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithGRPCConn(f.connMgr.GetConnection()))
	case ExporterConsole:
		return stdouttrace.New()
	default:
		return nil, fmt.Errorf("unsupported trace exporter: %s", f.exporter)
	}
}

func (f *otelExporterFactory) CreateMetricExporter(ctx context.Context) (sdkMetric.Exporter, error) {
	switch f.exporter {
	case ExporterGrpc:
		return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithGRPCConn(f.connMgr.GetConnection()))
	case ExporterConsole:
		return stdoutmetric.New()
	default:
		return nil, fmt.Errorf("unsupported metric exporter: %s", f.exporter)
	}
}

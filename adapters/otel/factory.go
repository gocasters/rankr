package otel

import (
	"fmt"
)

type AdapterFactory struct{}

func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{}
}

func (f *AdapterFactory) CreateAdapter(config Config) (OtelAdapter, error) {

	connMgr, err := f.createConnectionManager(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	exporterFactory := f.createExporterFactory(config.Exporter, connMgr)

	resourceBuilder := f.createResourceBuilder()

	tracerProvider, err := f.createTracerProvider(exporterFactory, resourceBuilder, config.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	metricProvider, err := f.createMetricProvider(exporterFactory, resourceBuilder, config.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric provider: %w", err)
	}

	propagationHandler := f.createPropagationHandler()

	return &compositeOtelAdapter{
		tracerProvider:     tracerProvider,
		metricProvider:     metricProvider,
		propagationHandler: propagationHandler,
		connectionManager:  connMgr,
		isConfigured:       true,
	}, nil
}

func (f *AdapterFactory) createConnectionManager(endpoint string) (ConnectionManager, error) {
	return newGrpcConnectionManager(endpoint)
}

func (f *AdapterFactory) createExporterFactory(exporterType Exporter, connMgr ConnectionManager) ExporterFactory {
	return newOtelExporterFactory(exporterType, connMgr)
}

func (f *AdapterFactory) createResourceBuilder() ResourceBuilder {
	return newDefaultResourceBuilder()
}

func (f *AdapterFactory) createTracerProvider(exporterFactory ExporterFactory, resourceBuilder ResourceBuilder, serviceName string) (TracerProvider, error) {
	return newOtelTracerProvider(exporterFactory, resourceBuilder, serviceName)
}

func (f *AdapterFactory) createMetricProvider(exporterFactory ExporterFactory, resourceBuilder ResourceBuilder, serviceName string) (MetricProvider, error) {
	return newOtelMetricProvider(exporterFactory, resourceBuilder, serviceName)
}

func (f *AdapterFactory) createPropagationHandler() PropagationHandler {
	return newOtelPropagationHandler()
}

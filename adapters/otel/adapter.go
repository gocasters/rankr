package otel

import (
	"context"
	"fmt"

	"github.com/gocasters/rankr/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type compositeOtelAdapter struct {
	tracerProvider     TracerProvider
	metricProvider     MetricProvider
	propagationHandler PropagationHandler
	connectionManager  ConnectionManager
	isConfigured       bool
}

func NewOtelAdapter(config Config) (OtelAdapter, error) {
	factory := NewAdapterFactory()
	return factory.CreateAdapter(config)
}

func (a *compositeOtelAdapter) NewTracer(name string, options ...trace.TracerOption) (trace.Tracer, error) {
	return a.tracerProvider.NewTracer(name, options...)
}

func (a *compositeOtelAdapter) NewMeter(name string, options ...metric.MeterOption) metric.Meter {
	return a.metricProvider.NewMeter(name, options...)
}

func (a *compositeOtelAdapter) GetCarrierFromContext(ctx context.Context) map[string]string {
	return a.propagationHandler.GetCarrierFromContext(ctx)
}

func (a *compositeOtelAdapter) GetContextFromCarrier(carrier map[string]string) context.Context {
	return a.propagationHandler.GetContextFromCarrier(carrier)
}

func (a *compositeOtelAdapter) IsConfigured() bool {
	return a.isConfigured
}

func (a *compositeOtelAdapter) AddFloat64Counter(ctx context.Context, meter metric.Meter, name, desc string, cv float64, options ...metric.Float64CounterOption) error {
	tracer, err := a.NewTracer("otel-adapter")
	if err != nil {
		logger.L().Error("failed to create tracer", err)
	}

	ctx, span := tracer.Start(ctx, "add-float64-counter")
	defer span.End()

	options = append(options, metric.WithDescription(desc))

	processedEventCounter, err := meter.Float64Counter(name, options...)
	if err != nil {
		span.AddEvent("error on create counter", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("counter_name", name),
			attribute.String("description", desc)))
		logger.L().Error("error on create counter",
			"error", err.Error(),
			"counter_name", name,
			"description", desc)
		return fmt.Errorf("failed to create counter %s: %w", name, err)
	}

	processedEventCounter.Add(ctx, cv)

	span.AddEvent("counter created and updated", trace.WithAttributes(
		attribute.String("counter_name", name),
		attribute.Float64("value", cv)))
	return nil
}

func (a *compositeOtelAdapter) Shutdown(ctx context.Context) error {
	var errs []error

	if err := a.tracerProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
	}

	if err := a.metricProvider.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown metric provider: %w", err))
	}

	if err := a.connectionManager.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close connection manager: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

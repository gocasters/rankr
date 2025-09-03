package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
	otelTrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OtelConfig struct {
	ServiceName string
	Endpoint    string
	UseConsole  bool
}

type OtelClient struct {
	tracer         otelTrace.Tracer
	meter          metric.Meter
	propagator     propagation.TextMapPropagator
	meterProvider  *sdkmetric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
	conn           *grpc.ClientConn
}

func NewOtelClient(config OtelConfig) (*OtelClient, error) {
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	client := &OtelClient{}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	if err := client.setupProviders(config, res); err != nil {
		return nil, fmt.Errorf("failed to setup providers: %w", err)
	}

	client.tracer = otel.Tracer(config.ServiceName)
	client.meter = otel.Meter(config.ServiceName)

	client.propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(client.propagator)

	return client, nil
}

func (c *OtelClient) setupProviders(config OtelConfig, res *resource.Resource) error {
	if config.UseConsole {
		return c.setupConsoleProviders(res)
	}
	return c.setupOTLPProviders(config, res)
}

func (c *OtelClient) setupConsoleProviders(res *resource.Resource) error {

	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return fmt.Errorf("failed to create console trace exporter: %w", err)
	}

	c.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(c.tracerProvider)

	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return fmt.Errorf("failed to create console metric exporter: %w", err)
	}

	c.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(c.meterProvider)

	return nil
}

func (c *OtelClient) setupOTLPProviders(config OtelConfig, res *resource.Resource) error {
	if config.Endpoint == "" {
		return fmt.Errorf("endpoint is required for OTLP exporters")
	}

	conn, err := grpc.Dial(config.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	c.conn = conn

	traceExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	c.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(c.tracerProvider)

	metricExporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	c.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(c.meterProvider)

	return nil
}

func (c *OtelClient) StartSpan(ctx context.Context, name string) (context.Context, otelTrace.Span) {
	return c.tracer.Start(ctx, name)
}

func (c *OtelClient) Counter(name, description string) (metric.Float64Counter, error) {
	return c.meter.Float64Counter(name, metric.WithDescription(description))
}

func (c *OtelClient) Histogram(name, description string) (metric.Float64Histogram, error) {
	return c.meter.Float64Histogram(name, metric.WithDescription(description))
}

func (c *OtelClient) InjectContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	c.propagator.Inject(ctx, carrier)
}

func (c *OtelClient) ExtractContext(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return c.propagator.Extract(ctx, carrier)
}

func (c *OtelClient) CounterWithSpan(ctx context.Context, name, description string, value float64, attrs ...attribute.KeyValue) error {
	ctx, span := c.StartSpan(ctx, "counter_operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("counter.name", name),
		attribute.Float64("counter.value", value),
	)
	span.SetAttributes(attrs...)

	counter, err := c.Counter(name, description)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return fmt.Errorf("failed to create counter: %w", err)
	}

	counter.Add(ctx, value, metric.WithAttributes(attrs...))
	span.AddEvent("counter updated")

	return nil
}

func (c *OtelClient) Shutdown(ctx context.Context) error {
	var errs []error

	if c.tracerProvider != nil {
		if err := c.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}

	if c.meterProvider != nil {
		if err := c.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("gRPC connection close: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

func QuickStart(serviceName string) (*OtelClient, error) {
	return NewOtelClient(OtelConfig{
		ServiceName: serviceName,
		UseConsole:  true,
	})
}

func QuickStartOTLP(serviceName, endpoint string) (*OtelClient, error) {
	return NewOtelClient(OtelConfig{
		ServiceName: serviceName,
		Endpoint:    endpoint,
		UseConsole:  false,
	})
}

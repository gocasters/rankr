package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocasters/rankr/adapters/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	fmt.Println("OpenTelemetry Adapter Example")

	fmt.Println("\nCreating adapter with Console exporter...")
	consoleAdapter := createConsoleAdapter()
	defer func(consoleAdapter otel.OtelAdapter, ctx context.Context) {
		err := consoleAdapter.Shutdown(ctx)
		if err != nil {
			log.Printf("Failed to shutdown console adapter: %v", err)
		}
	}(consoleAdapter, context.Background())

	fmt.Println("\nCreating adapter with OTLP gRPC exporter...")
	grpcAdapter := createGRPCAdapter()
	defer func(grpcAdapter otel.OtelAdapter, ctx context.Context) {
		err := grpcAdapter.Shutdown(ctx)
		if err != nil {
			log.Printf("Failed to shutdown grpc adapter: %v", err)
		}
	}(grpcAdapter, context.Background())

	fmt.Println("\nDemonstrating tracing...")
	demonstrateTracing(consoleAdapter)

	fmt.Println("\nDemonstrating metrics...")
	demonstrateMetrics(consoleAdapter)

	fmt.Println("\nDemonstrating context propagation...")
	demonstrateContextPropagation(consoleAdapter)

	fmt.Println("\nDemonstrating custom AddFloat64Counter...")
	demonstrateCustomCounter(consoleAdapter)

	fmt.Println("\nExample completed successfully!")
}

func createConsoleAdapter() otel.OtelAdapter {
	config := otel.Config{
		ServiceName: "rankr-example",
		Exporter:    otel.ExporterConsole,
	}

	adapter, err := otel.NewOtelAdapter(config)
	if err != nil {
		log.Fatalf("Failed to create console adapter: %v", err)
	}

	fmt.Printf(" Console adapter created and configured: %v\n", adapter.IsConfigured())
	return adapter
}

func createGRPCAdapter() otel.OtelAdapter {
	config := otel.Config{
		Endpoint:    "localhost:4317",
		ServiceName: "rankr-production",
		Exporter:    otel.ExporterGrpc,
	}

	adapter, err := otel.NewOtelAdapter(config)
	if err != nil {
		log.Printf("Warning: Failed to create gRPC adapter (this is expected if no OTLP collector is running): %v", err)

		return createConsoleAdapter()
	}

	fmt.Printf("gRPC adapter created and configured: %v\n", adapter.IsConfigured())
	return adapter
}

func demonstrateTracing(adapter otel.OtelAdapter) {
	tracer, err := adapter.NewTracer("example-tracer")
	if err != nil {
		log.Printf("Failed to create tracer: %v", err)
		return
	}

	ctx := context.Background()

	ctx, parentSpan := tracer.Start(ctx, "parent-operation")
	parentSpan.SetAttributes(
		attribute.String("operation.type", "example"),
		attribute.Int("operation.version", 1),
	)

	time.Sleep(10 * time.Millisecond)

	_, childSpan := tracer.Start(ctx, "child-operation")
	childSpan.SetAttributes(
		attribute.String("child.task", "processing"),
		attribute.Bool("child.success", true),
	)

	childSpan.AddEvent("Processing started")

	time.Sleep(5 * time.Millisecond)

	childSpan.AddEvent("Processing completed")

	childSpan.End()
	parentSpan.End()

	fmt.Println("Tracing example completed")
}

func demonstrateMetrics(adapter otel.OtelAdapter) {
	meter, _ := adapter.NewMeter("example-meter")

	requestCounter, err := meter.Float64Counter(
		"requests_total",
	)
	if err != nil {
		log.Printf("Failed to create counter: %v", err)
		return
	}

	requestDuration, err := meter.Float64Histogram(
		"request_duration_seconds",
	)
	if err != nil {
		log.Printf("Failed to create histogram: %v", err)
		return
	}

	ctx := context.Background()

	for i := 0; i < 5; i++ {

		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "GET"),
				attribute.String("status", "200"),
			),
		)

		duration := float64(100+i*50) / 1000.0
		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("endpoint", "/api/users"),
			),
		)

		time.Sleep(1 * time.Millisecond)
	}

	fmt.Println("Metrics example completed ")
}

func demonstrateContextPropagation(adapter otel.OtelAdapter) {

	tracer, err := adapter.NewTracer("propagation-tracer")
	if err != nil {
		log.Printf("Failed to create tracer for context propagation: %v", err)
		return
	}
	ctx, span := tracer.Start(context.Background(), "propagation-example")
	span.SetAttributes(attribute.String("example", "context-propagation"))

	carrier := adapter.GetCarrierFromContext(ctx)
	fmt.Printf("Extracted carrier from context: %v\n", carrier)

	newCtx := adapter.GetContextFromCarrier(carrier)

	_, newSpan := tracer.Start(newCtx, "propagated-operation")
	newSpan.SetAttributes(attribute.String("propagated", "true"))

	newSpan.End()
	span.End()

	fmt.Println("Context propagation example completed")
}

func demonstrateCustomCounter(adapter otel.OtelAdapter) {
	ctx := context.Background()
	meter, _ := adapter.NewMeter("custom-meter")

	err := adapter.AddFloat64Counter(ctx, meter, "custom_operations_total",
		"Total number of custom operations", 1.0)
	if err != nil {
		log.Printf("Failed to add custom counter: %v", err)
		return
	}

	err = adapter.AddFloat64Counter(ctx, meter, "custom_operations_total",
		"Total number of custom operations", 2.5)
	if err != nil {
		log.Printf("Failed to add custom counter with value 2.5: %v", err)
		return
	}

	err = adapter.AddFloat64Counter(ctx, meter, "error_operations_total",
		"Total number of error operations", 1.0)
	if err != nil {
		log.Printf("Failed to add error counter: %v", err)
		return
	}

	fmt.Println("Custom counter example completed")
}

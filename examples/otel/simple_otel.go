package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocasters/rankr/adapters/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	fmt.Println("Simple OpenTelemetry Example")
	fmt.Println("===========================")

	client, err := otel.QuickStart("simple-app")
	if err != nil {
		log.Fatalf("Failed to create OTel client: %v", err)
	}
	defer client.Shutdown(context.Background())

	fmt.Println("\nSimple Tracing:")
	simpleTracing(client)

	fmt.Println("\nSimple Metrics:")
	simpleMetrics(client)

	fmt.Println("\nCounter with Tracing:")
	counterWithTracing(client)

	fmt.Println("\nContext Propagation:")
	contextPropagation(client)

	fmt.Println("\nExample completed!")
}

func simpleTracing(client *otel.OtelClient) {
	ctx := context.Background()

	ctx, span := client.StartSpan(ctx, "user-operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", "123"),
		attribute.String("operation", "fetch_profile"),
	)

	time.Sleep(50 * time.Millisecond)

	span.AddEvent("processing completed")

	fmt.Println("Span created and completed")
}

func simpleMetrics(client *otel.OtelClient) {
	ctx := context.Background()

	requestCounter, err := client.Counter("requests_total", "Total number of requests")
	if err != nil {
		log.Printf("Failed to create counter: %v", err)
		return
	}

	duration, err := client.Histogram("request_duration", "Request duration in seconds")
	if err != nil {
		log.Printf("Failed to create histogram: %v", err)
		return
	}

	for i := 0; i < 3; i++ {
		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "GET"),
				attribute.String("endpoint", "/api/users"),
			))

		duration.Record(ctx, 0.1+float64(i)*0.05,
			metric.WithAttributes(
				attribute.String("endpoint", "/api/users"),
			))
	}

	fmt.Println("Metrics recorded")
}

func counterWithTracing(client *otel.OtelClient) {
	ctx := context.Background()

	err := client.CounterWithSpan(ctx, "operations_total", "Total operations", 1.0,
		attribute.String("operation", "data_processing"),
		attribute.String("status", "success"),
	)
	if err != nil {
		log.Printf("Failed to record counter with span: %v", err)
		return
	}

	fmt.Println("Counter recorded with automatic tracing")
}

func contextPropagation(client *otel.OtelClient) {

	ctx, span := client.StartSpan(context.Background(), "service-a")
	defer span.End()

	headers := make(propagation.MapCarrier)

	client.InjectContext(ctx, headers)
	fmt.Printf("Injected context into headers: %v\n", map[string]string(headers))

	newCtx := client.ExtractContext(context.Background(), headers)

	_, childSpan := client.StartSpan(newCtx, "service-b")
	childSpan.SetAttributes(attribute.String("service", "b"))
	childSpan.End()

	fmt.Println("Context propagated and child span created")
}

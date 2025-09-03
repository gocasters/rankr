package otel

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

func TestQuickStart(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	if client == nil {
		t.Error("Client should not be nil")
	}
}

func TestQuickStartOTLP(t *testing.T) {

	client, err := QuickStartOTLP("test-service", "localhost:4317")
	if err != nil {
		t.Logf("OTLP client creation failed (expected if no collector): %v", err)
		return
	}
	defer client.Shutdown(context.Background())

	if client == nil {
		t.Error("Client should not be nil")
	}
}

func TestOtelClient_StartSpan(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := client.StartSpan(ctx, "test-span")

	if span == nil {
		t.Error("Span should not be nil")
	}

	span.SetAttributes(attribute.String("test", "value"))
	span.End()
}

func TestOtelClient_Counter(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	counter, err := client.Counter("test_counter", "Test counter")
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	if counter == nil {
		t.Error("Counter should not be nil")
	}

	ctx := context.Background()
	counter.Add(ctx, 1.0, metric.WithAttributes(attribute.String("test", "value")))
}

func TestOtelClient_Histogram(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	histogram, err := client.Histogram("test_histogram", "Test histogram")
	if err != nil {
		t.Fatalf("Failed to create histogram: %v", err)
	}

	if histogram == nil {
		t.Error("Histogram should not be nil")
	}

	ctx := context.Background()
	histogram.Record(ctx, 0.5, metric.WithAttributes(attribute.String("test", "value")))
}

func TestOtelClient_CounterWithSpan(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	ctx := context.Background()
	err = client.CounterWithSpan(ctx, "test_counter_with_span", "Test counter with span", 1.0,
		attribute.String("operation", "test"))

	if err != nil {
		t.Fatalf("Failed to record counter with span: %v", err)
	}
}

func TestOtelClient_ContextPropagation(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	ctx, span := client.StartSpan(context.Background(), "parent-span")
	defer span.End()

	carrier := make(propagation.MapCarrier)
	client.InjectContext(ctx, carrier)

	if len(carrier) == 0 {
		t.Error("Carrier should contain injected context")
	}

	newCtx := client.ExtractContext(context.Background(), carrier)
	if newCtx == nil {
		t.Error("Extracted context should not be nil")
	}

	_, childSpan := client.StartSpan(newCtx, "child-span")
	childSpan.End()
}

func TestOtelClient_Shutdown(t *testing.T) {
	client, err := QuickStart("test-service")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown client: %v", err)
	}
}

func TestOtelConfig_Validation(t *testing.T) {

	_, err := NewOtelClient(OtelConfig{})
	if err == nil {
		t.Error("Should fail with empty service name")
	}

	_, err = NewOtelClient(OtelConfig{
		ServiceName: "test",
		UseConsole:  false,
		Endpoint:    "",
	})
	if err == nil {
		t.Error("Should fail with missing endpoint for OTLP")
	}
}

func TestIntegration(t *testing.T) {
	client, err := QuickStart("integration-test")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Shutdown(context.Background())

	ctx := context.Background()

	ctx, span := client.StartSpan(ctx, "integration-operation")
	defer span.End()

	err = client.CounterWithSpan(ctx, "integration_operations", "Integration operations", 1.0,
		attribute.String("test", "integration"))
	if err != nil {
		t.Fatalf("Failed to record counter: %v", err)
	}

	carrier := make(propagation.MapCarrier)
	client.InjectContext(ctx, carrier)

	newCtx := client.ExtractContext(context.Background(), carrier)
	_, childSpan := client.StartSpan(newCtx, "child-operation")
	childSpan.End()
}

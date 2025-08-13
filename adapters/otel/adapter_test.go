package otel

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	noopTrace "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
)

type mockTracerProvider struct {
	shutdownErr  error
	startedSpans []*mockSpan
}

func (m *mockTracerProvider) NewTracer(name string, options ...trace.TracerOption) (trace.Tracer, error) {
	return &mockTracer{provider: m}, nil
}

func (m *mockTracerProvider) Shutdown(ctx context.Context) error {
	return m.shutdownErr
}

type mockMetricProvider struct {
	shutdownErr     error
	createdCounters map[string]*mockFloat64Counter
}

func (m *mockMetricProvider) NewMeter(name string, options ...metric.MeterOption) metric.Meter {
	if m.createdCounters == nil {
		m.createdCounters = make(map[string]*mockFloat64Counter)
	}
	return &mockMeter{provider: m}
}

func (m *mockMetricProvider) Shutdown(ctx context.Context) error {
	return m.shutdownErr
}

type mockPropagationHandler struct {
	carrier map[string]string
	context context.Context
}

func (m *mockPropagationHandler) GetCarrierFromContext(ctx context.Context) map[string]string {
	if m.carrier != nil {
		return m.carrier
	}
	return map[string]string{"test": "value"}
}

func (m *mockPropagationHandler) GetContextFromCarrier(carrier map[string]string) context.Context {
	if m.context != nil {
		return m.context
	}
	return context.Background()
}

type mockConnectionManager struct {
	closeErr error
}

func (m *mockConnectionManager) GetConnection() *grpc.ClientConn {
	return nil
}

func (m *mockConnectionManager) Close() error {
	return m.closeErr
}

type mockTracer struct {
	noopTrace.Tracer
	provider *mockTracerProvider
}

func (m *mockTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	span := &mockSpan{
		Span: noopTrace.Span{},
		name: spanName,
	}
	if m.provider != nil {
		m.provider.startedSpans = append(m.provider.startedSpans, span)
	}
	return ctx, span
}

type mockSpan struct {
	noopTrace.Span
	name       string
	events     []string
	ended      bool
	attributes []attribute.KeyValue
}

func (m *mockSpan) End(options ...trace.SpanEndOption) {
	m.ended = true
}

func (m *mockSpan) AddEvent(name string, options ...trace.EventOption) {
	m.events = append(m.events, name)
}

func (m *mockSpan) SetName(name string) {
	m.name = name
}

func (m *mockSpan) SetAttributes(kv ...attribute.KeyValue) {
	m.attributes = append(m.attributes, kv...)
}

func (m *mockSpan) RecordError(err error, options ...trace.EventOption) {
	m.events = append(m.events, "error: "+err.Error())
}

type mockMeter struct {
	noop.Meter
	provider *mockMetricProvider
}

func (m *mockMeter) Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	counter := &mockFloat64Counter{
		Float64Counter: noop.Float64Counter{},
		name:           name,
	}
	if m.provider != nil {
		m.provider.createdCounters[name] = counter
	}
	return counter, nil
}

type mockFloat64Counter struct {
	noop.Float64Counter
	name  string
	value float64
}

func (m *mockFloat64Counter) Add(ctx context.Context, incr float64, options ...metric.AddOption) {
	m.value += incr
}

func TestCompositeOtelAdapter_NewTracer(t *testing.T) {
	adapter := &compositeOtelAdapter{
		tracerProvider: &mockTracerProvider{},
		isConfigured:   true,
	}

	tracer, _ := adapter.NewTracer("test-tracer")
	if tracer == nil {
		t.Error("NewTracer should return a non-nil tracer")
	}
}

func TestCompositeOtelAdapter_NewMeter(t *testing.T) {
	adapter := &compositeOtelAdapter{
		metricProvider: &mockMetricProvider{},
		isConfigured:   true,
	}

	meter, _ := adapter.NewMeter("test-meter")
	if meter == nil {
		t.Error("NewMeter should return a non-nil meter")
	}
}

func TestCompositeOtelAdapter_GetCarrierFromContext(t *testing.T) {
	expectedCarrier := map[string]string{"trace-id": "123", "span-id": "456"}
	adapter := &compositeOtelAdapter{
		propagationHandler: &mockPropagationHandler{carrier: expectedCarrier},
		isConfigured:       true,
	}

	carrier := adapter.GetCarrierFromContext(context.Background())
	if len(carrier) != len(expectedCarrier) {
		t.Errorf("Expected carrier length %d, got %d", len(expectedCarrier), len(carrier))
	}
	for k, v := range expectedCarrier {
		if carrier[k] != v {
			t.Errorf("Expected carrier[%s] = %s, got %s", k, v, carrier[k])
		}
	}
}

func TestCompositeOtelAdapter_GetContextFromCarrier(t *testing.T) {
	expectedCtx := context.WithValue(context.Background(), "test", "value")
	adapter := &compositeOtelAdapter{
		propagationHandler: &mockPropagationHandler{context: expectedCtx},
		isConfigured:       true,
	}

	carrier := map[string]string{"trace-id": "123"}
	ctx := adapter.GetContextFromCarrier(carrier)
	if ctx != expectedCtx {
		t.Error("GetContextFromCarrier should return the context from the handler")
	}
}

func TestCompositeOtelAdapter_IsConfigured(t *testing.T) {
	tests := []struct {
		name         string
		isConfigured bool
	}{
		{"configured adapter", true},
		{"unconfigured adapter", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &compositeOtelAdapter{
				isConfigured: tt.isConfigured,
			}

			if adapter.IsConfigured() != tt.isConfigured {
				t.Errorf("IsConfigured() = %v, want %v", adapter.IsConfigured(), tt.isConfigured)
			}
		})
	}
}

func TestCompositeOtelAdapter_AddFloat64Counter(t *testing.T) {
	mockTracerProvider := &mockTracerProvider{}
	mockMetricProvider := &mockMetricProvider{createdCounters: make(map[string]*mockFloat64Counter)}

	adapter := &compositeOtelAdapter{
		tracerProvider: mockTracerProvider,
		metricProvider: mockMetricProvider,
		isConfigured:   true,
	}

	ctx := context.Background()
	meter, _ := adapter.NewMeter("test-meter")

	adapter.AddFloat64Counter(ctx, meter, "test_counter", "Test counter description", 1.5)

	if len(mockTracerProvider.startedSpans) != 1 {
		t.Errorf("Expected 1 span to be created, got %d", len(mockTracerProvider.startedSpans))
	}

	span := mockTracerProvider.startedSpans[0]
	if span.name != "add-float64-counter" {
		t.Errorf("Expected span name 'add-float64-counter', got '%s'", span.name)
	}

	if !span.ended {
		t.Error("Span should be ended after AddFloat64Counter completes")
	}

	if mockMetricProvider.createdCounters["test_counter"] == nil {
		t.Error("Counter should be created")
	} else if mockMetricProvider.createdCounters["test_counter"].value != 1.5 {
		t.Errorf("Expected counter value 1.5, got %f", mockMetricProvider.createdCounters["test_counter"].value)
	}

	expectedEvents := []string{"counter created and updated"}
	if len(span.events) != len(expectedEvents) {
		t.Errorf("Expected %d events, got %d", len(expectedEvents), len(span.events))
	}
}

func TestCompositeOtelAdapter_Shutdown(t *testing.T) {
	tests := []struct {
		name               string
		tracerShutdownErr  error
		metricShutdownErr  error
		connectionCloseErr error
		expectError        bool
	}{
		{
			name:        "successful shutdown",
			expectError: false,
		},
		{
			name:              "tracer shutdown error",
			tracerShutdownErr: errors.New("tracer shutdown failed"),
			expectError:       true,
		},
		{
			name:              "metric shutdown error",
			metricShutdownErr: errors.New("metric shutdown failed"),
			expectError:       true,
		},
		{
			name:               "connection close error",
			connectionCloseErr: errors.New("connection close failed"),
			expectError:        true,
		},
		{
			name:               "multiple errors",
			tracerShutdownErr:  errors.New("tracer error"),
			metricShutdownErr:  errors.New("metric error"),
			connectionCloseErr: errors.New("connection error"),
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &compositeOtelAdapter{
				tracerProvider:    &mockTracerProvider{shutdownErr: tt.tracerShutdownErr},
				metricProvider:    &mockMetricProvider{shutdownErr: tt.metricShutdownErr},
				connectionManager: &mockConnectionManager{closeErr: tt.connectionCloseErr},
				isConfigured:      true,
			}

			err := adapter.Shutdown(context.Background())

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestNewOtelAdapter(t *testing.T) {

	config := Config{
		Endpoint:    "localhost:4317",
		ServiceName: "test-service",
		Exporter:    ExporterConsole,
	}

	adapter, err := NewOtelAdapter(config)
	if err != nil {
		t.Fatalf("NewOtelAdapter failed: %v", err)
	}

	if adapter == nil {
		t.Error("NewOtelAdapter should return a non-nil adapter")
	}

	if !adapter.IsConfigured() {
		t.Error("Adapter should be configured after creation")
	}

	tracer, _ := adapter.NewTracer("test")
	if tracer == nil {
		t.Error("NewTracer should return a non-nil tracer")
	}

	meter := adapter.NewMeter("test")
	if meter == nil {
		t.Error("NewMeter should return a non-nil meter")
	}

	if err := adapter.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

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

func (m *mockTracerProvider) NewTracer(name string, options ...trace.TracerOption) trace.Tracer {
	return &mockTracer{provider: m}
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

	tracer := adapter.NewTracer("test-tracer")
	if tracer == nil {
		t.Error("NewTracer should return a non-nil tracer")
	}
}

func TestCompositeOtelAdapter_NewMeter(t *testing.T) {
	adapter := &compositeOtelAdapter{
		metricProvider: &mockMetricProvider{},
		isConfigured:   true,
	}

	meter := adapter.NewMeter("test-meter")
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
	meter := adapter.NewMeter("test-meter")

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

	tracer := adapter.NewTracer("test")
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

// Test factory creation and adapter initialization
func TestAdapterFactory_CreateAdapter(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid console config",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				Exporter:    ExporterConsole,
			},
			wantError: false,
		},
		{
			name: "valid grpc config",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				Exporter:    ExporterGrpc,
			},
			wantError: false,
		},
		{
			name: "invalid exporter",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				Exporter:    "invalid",
			},
			wantError: true,
			errorMsg:  "unsupported",
		},
		{
			name: "invalid endpoint for grpc",
			config: Config{
				Endpoint:    "invalid:endpoint:format",
				ServiceName: "test-service",
				Exporter:    ExporterGrpc,
			},
			wantError: true,
			errorMsg:  "connection manager",
		},
		{
			name: "empty service name",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "",
				Exporter:    ExporterConsole,
			},
			wantError: false, // Should use empty service name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewAdapterFactory()
			adapter, err := factory.CreateAdapter(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s' but got none", tt.errorMsg)
				} else if tt.errorMsg != "" && !containsIgnoreCase(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if adapter == nil {
					t.Error("Expected non-nil adapter")
				} else {
					if !adapter.IsConfigured() {
						t.Error("Adapter should be configured")
					}
				}
			}

			// Clean up if adapter was created
			if adapter != nil {
				_ = adapter.Shutdown(context.Background())
			}
		})
	}
}

// Test NewAdapterFactory
func TestNewAdapterFactory(t *testing.T) {
	factory := NewAdapterFactory()
	if factory == nil {
		t.Error("NewAdapterFactory should return non-nil factory")
	}

	// Test that we can create multiple factories
	factory2 := NewAdapterFactory()
	if factory2 == nil {
		t.Error("Second NewAdapterFactory should return non-nil factory")
	}

	// Factories should be independent
	if factory == factory2 {
		t.Error("Different factory instances should be independent")
	}
}

// Test error scenarios in AddFloat64Counter with various edge cases
func TestCompositeOtelAdapter_AddFloat64Counter_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		counterName string
		description string
		value       float64
		setupMocks  func() (*mockTracerProvider, MetricProvider)
		expectError bool
	}{
		{
			name:        "normal operation",
			counterName: "normal_counter",
			description: "Normal counter",
			value:       1.0,
			setupMocks: func() (*mockTracerProvider, MetricProvider) {
				return &mockTracerProvider{}, &mockMetricProvider{createdCounters: make(map[string]*mockFloat64Counter)}
			},
			expectError: false,
		},
		{
			name:        "counter creation fails",
			counterName: "failing_counter",
			description: "Failing counter",
			value:       1.0,
			setupMocks: func() (*mockTracerProvider, MetricProvider) {
				return &mockTracerProvider{}, &mockMetricProviderWithError{
					mockMetricProvider: &mockMetricProvider{createdCounters: make(map[string]*mockFloat64Counter)},
					counterError:       errors.New("counter creation failed"),
				}
			},
			expectError: true,
		},
		{
			name:        "infinity value",
			counterName: "infinity_counter",
			description: "Infinity counter",
			value:       float64(1) / float64(0), // +Inf
			setupMocks: func() (*mockTracerProvider, MetricProvider) {
				return &mockTracerProvider{}, &mockMetricProvider{createdCounters: make(map[string]*mockFloat64Counter)}
			},
			expectError: false,
		},
		{
			name:        "negative infinity value",
			counterName: "neg_infinity_counter",
			description: "Negative infinity counter",
			value:       float64(-1) / float64(0), // -Inf
			setupMocks: func() (*mockTracerProvider, MetricProvider) {
				return &mockTracerProvider{}, &mockMetricProvider{createdCounters: make(map[string]*mockFloat64Counter)}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTracerProvider, mockMetricProvider := tt.setupMocks()

			adapter := &compositeOtelAdapter{
				tracerProvider: mockTracerProvider,
				metricProvider: mockMetricProvider,
				isConfigured:   true,
			}

			ctx := context.Background()
			meter := adapter.NewMeter("test-meter")

			adapter.AddFloat64Counter(ctx, meter, tt.counterName, tt.description, tt.value)

			// Verify span was created
			if len(mockTracerProvider.startedSpans) != 1 {
				t.Errorf("Expected 1 span, got %d", len(mockTracerProvider.startedSpans))
			}

			span := mockTracerProvider.startedSpans[0]
			if !span.ended {
				t.Error("Span should be ended")
			}

			if tt.expectError {
				// Should have error event
				hasErrorEvent := false
				for _, event := range span.events {
					if event == "error on create counter" {
						hasErrorEvent = true
						break
					}
				}
				if !hasErrorEvent {
					t.Error("Expected error event in span when counter creation fails")
				}
			} else {
				// Should have success event
				hasSuccessEvent := false
				for _, event := range span.events {
					if event == "counter created and updated" {
						hasSuccessEvent = true
						break
					}
				}
				if !hasSuccessEvent {
					t.Error("Expected success event in span when counter creation succeeds")
				}
			}
		})
	}
}

// Test context propagation with various context states
func TestCompositeOtelAdapter_ContextPropagation_States(t *testing.T) {
	tests := []struct {
		name             string
		inputContext     context.Context
		inputCarrier     map[string]string
		handlerCarrier   map[string]string
		handlerContext   context.Context
		expectedCarrier  map[string]string
		expectedContext  context.Context
	}{
		{
			name:            "normal context",
			inputContext:    context.Background(),
			handlerCarrier:  map[string]string{"trace-id": "123"},
			expectedCarrier: map[string]string{"trace-id": "123"},
		},
		{
			name:         "context with value",
			inputContext: context.WithValue(context.Background(), "key", "value"),
			handlerCarrier: map[string]string{"trace-id": "456", "span-id": "789"},
			expectedCarrier: map[string]string{"trace-id": "456", "span-id": "789"},
		},
		{
			name:         "context with deadline",
			inputContext: func() context.Context { 
				ctx, _ := context.WithTimeout(context.Background(), 0)
				return ctx
			}(),
			handlerCarrier:  map[string]string{"custom": "header"},
			expectedCarrier: map[string]string{"custom": "header"},
		},
		{
			name:            "nil carrier from handler",
			inputContext:    context.Background(),
			handlerCarrier:  nil,
			expectedCarrier: map[string]string{"test": "value"}, // Default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &mockPropagationHandler{
				carrier: tt.handlerCarrier,
				context: tt.handlerContext,
			}

			adapter := &compositeOtelAdapter{
				propagationHandler: handler,
				isConfigured:       true,
			}

			// Test GetCarrierFromContext
			carrier := adapter.GetCarrierFromContext(tt.inputContext)
			if len(carrier) != len(tt.expectedCarrier) {
				t.Errorf("Expected carrier length %d, got %d", len(tt.expectedCarrier), len(carrier))
			}
			for k, expectedV := range tt.expectedCarrier {
				if actualV := carrier[k]; actualV != expectedV {
					t.Errorf("Expected carrier[%s] = %s, got %s", k, expectedV, actualV)
				}
			}

			// Test GetContextFromCarrier
			resultCtx := adapter.GetContextFromCarrier(tt.inputCarrier)
			if resultCtx == nil {
				t.Error("GetContextFromCarrier should never return nil")
			}
		})
	}
}

// Test adapter with nil components
func TestCompositeOtelAdapter_NilComponents(t *testing.T) {
	tests := []struct {
		name              string
		tracerProvider    TracerProvider
		metricProvider    MetricProvider
		propagationHandler PropagationHandler
		connectionManager ConnectionManager
		shouldPanic       bool
		operation         string
	}{
		{
			name:           "nil tracer provider",
			tracerProvider: nil,
			metricProvider: &mockMetricProvider{},
			propagationHandler: &mockPropagationHandler{},
			connectionManager: &mockConnectionManager{},
			shouldPanic:    true,
			operation:      "NewTracer",
		},
		{
			name:           "nil metric provider", 
			tracerProvider: &mockTracerProvider{},
			metricProvider: nil,
			propagationHandler: &mockPropagationHandler{},
			connectionManager: &mockConnectionManager{},
			shouldPanic:    true,
			operation:      "NewMeter",
		},
		{
			name:           "nil propagation handler",
			tracerProvider: &mockTracerProvider{},
			metricProvider: &mockMetricProvider{},
			propagationHandler: nil,
			connectionManager: &mockConnectionManager{},
			shouldPanic:    true,
			operation:      "GetCarrierFromContext",
		},
		{
			name:           "all nil providers",
			tracerProvider: nil,
			metricProvider: nil,
			propagationHandler: nil,
			connectionManager: nil,
			shouldPanic:    false, // Shutdown should handle nil gracefully
			operation:      "Shutdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &compositeOtelAdapter{
				tracerProvider:     tt.tracerProvider,
				metricProvider:     tt.metricProvider,
				propagationHandler: tt.propagationHandler,
				connectionManager:  tt.connectionManager,
				isConfigured:       true,
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Operation %s should not panic with nil components", tt.operation)
					}
				} else if tt.shouldPanic {
					t.Errorf("Operation %s should panic with nil components", tt.operation)
				}
			}()

			switch tt.operation {
			case "NewTracer":
				_ = adapter.NewTracer("test")
			case "NewMeter":
				_ = adapter.NewMeter("test")
			case "GetCarrierFromContext":
				_ = adapter.GetCarrierFromContext(context.Background())
			case "Shutdown":
				_ = adapter.Shutdown(context.Background())
			}
		})
	}
}

// Test configuration validation
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config with console exporter",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				Exporter:    ExporterConsole,
			},
			valid: true,
		},
		{
			name: "valid config with grpc exporter",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				Exporter:    ExporterGrpc,
			},
			valid: true,
		},
		{
			name: "config with custom exporter string",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "test-service",
				Exporter:    "custom",
			},
			valid: false,
		},
		{
			name: "empty endpoint with console exporter",
			config: Config{
				Endpoint:    "",
				ServiceName: "test-service",
				Exporter:    ExporterConsole,
			},
			valid: true, // Console doesn't need endpoint
		},
		{
			name: "empty service name",
			config: Config{
				Endpoint:    "localhost:4317",
				ServiceName: "",
				Exporter:    ExporterConsole,
			},
			valid: true, // Empty service name should be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewOtelAdapter(tt.config)
			
			if tt.valid {
				if err != nil {
					t.Errorf("Expected valid config to succeed, got error: %v", err)
				}
				if adapter != nil {
					defer adapter.Shutdown(context.Background())
				}
			} else {
				if err == nil {
					t.Error("Expected invalid config to fail")
					if adapter != nil {
						defer adapter.Shutdown(context.Background())
					}
				}
			}
		})
	}
}

// Test constants and exporter types
func TestExporterConstants(t *testing.T) {
	tests := []struct {
		name     string
		exporter Exporter
		expected string
	}{
		{
			name:     "grpc exporter constant",
			exporter: ExporterGrpc,
			expected: "grpc",
		},
		{
			name:     "console exporter constant",
			exporter: ExporterConsole,
			expected: "console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.exporter) != tt.expected {
				t.Errorf("Expected exporter constant %s, got %s", tt.expected, string(tt.exporter))
			}
		})
	}
}

// Test AddFloat64Counter with various metric options
func TestCompositeOtelAdapter_AddFloat64Counter_WithMetricOptions(t *testing.T) {
	mockTracerProvider := &mockTracerProvider{}
	mockMetricProvider := &mockMetricProvider{createdCounters: make(map[string]*mockFloat64Counter)}

	adapter := &compositeOtelAdapter{
		tracerProvider: mockTracerProvider,
		metricProvider: mockMetricProvider,
		isConfigured:   true,
	}

	ctx := context.Background()
	meter := adapter.NewMeter("options-test-meter")

	// Test with metric options (though our mock doesn't use them)
	counterOptions := []metric.Float64CounterOption{
		metric.WithUnit("requests"),
	}

	adapter.AddFloat64Counter(ctx, meter, "options_counter", "Counter with options", 5.0, counterOptions...)

	// Verify basic functionality still works
	if len(mockTracerProvider.startedSpans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(mockTracerProvider.startedSpans))
	}

	counter := mockMetricProvider.createdCounters["options_counter"]
	if counter == nil {
		t.Fatal("Counter should have been created")
	}

	if counter.value != 5.0 {
		t.Errorf("Expected counter value 5.0, got %f", counter.value)
	}
}

// Test span attributes and events in detail
func TestMockSpan_AttributesAndEvents(t *testing.T) {
	span := &mockSpan{}

	// Test initial state
	if span.ended {
		t.Error("Span should not be ended initially")
	}
	if len(span.events) != 0 {
		t.Error("Span should have no events initially")
	}
	if len(span.attributes) != 0 {
		t.Error("Span should have no attributes initially")
	}

	// Test event recording
	span.AddEvent("first-event")
	span.AddEvent("second-event")
	if len(span.events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(span.events))
	}

	// Test error recording
	testErr := errors.New("test error message")
	span.RecordError(testErr)
	if len(span.events) != 3 {
		t.Errorf("Expected 3 events after RecordError, got %d", len(span.events))
	}
	if span.events[2] != "error: test error message" {
		t.Errorf("Expected error event, got %s", span.events[2])
	}

	// Test attribute setting
	attrs := []attribute.KeyValue{
		attribute.String("string_attr", "value"),
		attribute.Int("int_attr", 42),
		attribute.Float64("float_attr", 3.14),
		attribute.Bool("bool_attr", true),
	}
	span.SetAttributes(attrs...)
	if len(span.attributes) != 4 {
		t.Errorf("Expected 4 attributes, got %d", len(span.attributes))
	}

	// Test name setting
	originalName := span.name
	span.SetName("new-span-name")
	if span.name == originalName {
		t.Error("Span name should have changed")
	}
	if span.name != "new-span-name" {
		t.Errorf("Expected span name 'new-span-name', got %s", span.name)
	}

	// Test ending
	span.End()
	if !span.ended {
		t.Error("Span should be ended after End() call")
	}
}

// Test mock counter behavior
func TestMockFloat64Counter_Behavior(t *testing.T) {
	counter := &mockFloat64Counter{}

	// Test initial value
	if counter.value != 0.0 {
		t.Errorf("Expected initial value 0.0, got %f", counter.value)
	}

	// Test adding values
	testValues := []float64{1.0, 2.5, -0.5, 10.7}
	expectedSum := 0.0

	for _, value := range testValues {
		counter.Add(context.Background(), value)
		expectedSum += value
	}

	if counter.value != expectedSum {
		t.Errorf("Expected counter value %f, got %f", expectedSum, counter.value)
	}

	// Test with context variations
	ctx, cancel := context.WithCancel(context.Background())
	counter.Add(ctx, 1.0)
	expectedSum += 1.0

	cancel()
	counter.Add(ctx, 2.0) // Should still work with cancelled context
	expectedSum += 2.0

	if counter.value != expectedSum {
		t.Errorf("Expected counter value %f after context operations, got %f", expectedSum, counter.value)
	}
}

// Test adapter interface compliance
func TestOtelAdapter_InterfaceCompliance(t *testing.T) {
	config := Config{
		Endpoint:    "localhost:4317",
		ServiceName: "interface-test",
		Exporter:    ExporterConsole,
	}

	adapter, err := NewOtelAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer adapter.Shutdown(context.Background())

	// Verify adapter implements all interface methods
	var _ TracerProvider = adapter
	var _ MetricProvider = adapter  
	var _ PropagationHandler = adapter
	var _ OtelAdapter = adapter

	// Test that all methods can be called
	ctx := context.Background()

	tracer := adapter.NewTracer("interface-test-tracer")
	if tracer == nil {
		t.Error("NewTracer should return non-nil tracer")
	}

	meter := adapter.NewMeter("interface-test-meter")  
	if meter == nil {
		t.Error("NewMeter should return non-nil meter")
	}

	carrier := adapter.GetCarrierFromContext(ctx)
	if carrier == nil {
		t.Error("GetCarrierFromContext should return non-nil carrier")
	}

	resultCtx := adapter.GetContextFromCarrier(map[string]string{"test": "value"})
	if resultCtx == nil {
		t.Error("GetContextFromCarrier should return non-nil context")
	}

	if !adapter.IsConfigured() {
		t.Error("IsConfigured should return true for properly created adapter")
	}

	// Test AddFloat64Counter (should not panic)
	adapter.AddFloat64Counter(ctx, meter, "interface_counter", "Interface test", 1.0)

	// Test Shutdown
	if err := adapter.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown should succeed, got error: %v", err)
	}
}

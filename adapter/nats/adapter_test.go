package nats

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/stretchr/testify/assert"
)

var validConfig = Config{
	URL:            "nats://localhost:4222",
	ClientID:       "test-client",
	DurableName:    "test-durable",
	QueueGroup:     "test-queue",
	AckWaitTimeout: 30 * time.Second,
	MaxInflight:    1024,
	ConnectTimeout: 5 * time.Second,
	ReconnectWait:  2 * time.Second,
	MaxReconnects:  -1,
	PingInterval:   2 * time.Minute,
	MaxPingsOut:    2,
	AllowReconnect: true,
	UseJetStream:   false,
}

func TestConfigValidate_Valid(t *testing.T) {
	vErrors := validConfig.Validate()
	assert.Empty(t, vErrors)
}

func TestConfigValidate_Invalid(t *testing.T) {
	testCases := []struct {
		name           string
		config         Config
		expectedErrors []string
	}{
		{
			name: "empty URL",
			config: Config{
				URL:            "",
				ClientID:       "test-client",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"url"},
		},
		{
			name: "empty client ID",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"client_id"},
		},
		{
			name: "invalid connect timeout",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "test-client",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 0,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"connect_timeout"},
		},
		{
			name: "invalid ack wait timeout",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "test-client",
				AckWaitTimeout: 0,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"ack_wait_timeout"},
		},
		{
			name: "invalid max inflight",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "test-client",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    0,
				ConnectTimeout: 5 * time.Second,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"max_inflight"},
		},
		{
			name: "invalid max reconnects",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "test-client",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				MaxReconnects:  -2,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"max_reconnects"},
		},
		{
			name: "invalid ping interval",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "test-client",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				PingInterval:   0,
				MaxPingsOut:    2,
			},
			expectedErrors: []string{"ping_interval"},
		},
		{
			name: "invalid max pings out",
			config: Config{
				URL:            "nats://localhost:4222",
				ClientID:       "test-client",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    0,
			},
			expectedErrors: []string{"max_pings_out"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vErrors := tc.config.Validate()
			assert.NotEmpty(t, vErrors)

			for _, expectedError := range tc.expectedErrors {
				_, exists := vErrors[expectedError]
				assert.True(t, exists, "Expected error for field: %s", expectedError)
			}
		})
	}
}

func TestConfigSetDefaults(t *testing.T) {
	config := Config{}
	config.SetDefaults()

	assert.Equal(t, "nats://127.0.0.1:4222", config.URL)
	assert.Equal(t, 30*time.Second, config.AckWaitTimeout)
	assert.Equal(t, 1024, config.MaxInflight)
	assert.Equal(t, 5*time.Second, config.ConnectTimeout)
	assert.Equal(t, 2*time.Second, config.ReconnectWait)
	assert.Equal(t, -1, config.MaxReconnects)
	assert.Equal(t, 2*time.Minute, config.PingInterval)
	assert.Equal(t, 2, config.MaxPingsOut)
	assert.True(t, config.AllowReconnect)
}

func TestConfigSetDefaults_PartialConfig(t *testing.T) {
	config := Config{
		URL:      "nats://custom:4223",
		ClientID: "custom-client",
	}
	config.SetDefaults()

	assert.Equal(t, "nats://custom:4223", config.URL)
	assert.Equal(t, "custom-client", config.ClientID)

	assert.Equal(t, 30*time.Second, config.AckWaitTimeout)
	assert.Equal(t, 1024, config.MaxInflight)
	assert.Equal(t, 5*time.Second, config.ConnectTimeout)
	assert.Equal(t, 2*time.Second, config.ReconnectWait)
	assert.Equal(t, -1, config.MaxReconnects)
	assert.Equal(t, 2*time.Minute, config.PingInterval)
	assert.Equal(t, 2, config.MaxPingsOut)
	assert.True(t, config.AllowReconnect)
}

func TestFormatValidationErrors(t *testing.T) {
	testCases := []struct {
		name           string
		errors         map[string]error
		expectedResult string
	}{
		{
			name:           "empty errors",
			errors:         map[string]error{},
			expectedResult: "",
		},
		{
			name: "single error",
			errors: map[string]error{
				"url": assert.AnError,
			},
			expectedResult: "validation errors: url: assert.AnError general error for testing",
		},
		{
			name: "multiple errors",
			errors: map[string]error{
				"url":       assert.AnError,
				"client_id": assert.AnError,
			},
			expectedResult: "validation errors: ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatValidationErrors(tc.errors)

			if tc.name == "empty errors" {
				assert.Equal(t, tc.expectedResult, result)
			} else if tc.name == "single error" {
				assert.Contains(t, result, "validation errors:")
				assert.Contains(t, result, "url:")
				assert.Contains(t, result, "assert.AnError")
			} else {
				assert.Contains(t, result, "validation errors:")
			}
		})
	}
}

func TestNew_NilContext(t *testing.T) {
	logger := watermill.NewStdLogger(false, false)

	adapter, err := New(nil, validConfig, logger)

	assert.Nil(t, adapter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
}

func TestNew_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	logger := watermill.NewStdLogger(false, false)

	invalidConfig := Config{
		URL:            "nats://localhost:4222",
		ClientID:       "",
		AckWaitTimeout: 30 * time.Second,
		MaxInflight:    1024,
		ConnectTimeout: 5 * time.Second,
		PingInterval:   2 * time.Minute,
		MaxPingsOut:    2,
	}

	adapter, err := New(ctx, invalidConfig, logger)

	assert.Nil(t, adapter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid NATS configuration")
}

func TestNew_NilLogger(t *testing.T) {
	ctx := context.Background()

	testConfig := validConfig
	testConfig.URL = "nats://invalid-host:4222"
	testConfig.ConnectTimeout = 50 * time.Millisecond
	testConfig.AllowReconnect = false

	adapter, err := New(ctx, testConfig, nil)

	assert.Nil(t, adapter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to NATS")
}

func TestAdapter_ConfigMethod(t *testing.T) {
	adapter := &Adapter{
		config: validConfig,
	}

	config := adapter.Config()
	assert.Equal(t, validConfig, config)
}

func TestAdapter_IsConnected_NilConnection(t *testing.T) {
	adapter := &Adapter{
		conn: nil,
	}

	assert.False(t, adapter.IsConnected())
}

func TestAdapter_Status_NilConnection(t *testing.T) {
	adapter := &Adapter{
		conn: nil,
	}

	status := adapter.Status()
	assert.Equal(t, status.String(), "DISCONNECTED")
}

func TestAdapter_Publish_NilPublisher(t *testing.T) {
	adapter := &Adapter{
		publisher: nil,
	}

	err := adapter.Publish("test.topic", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publisher is not initialized")
}

func TestAdapter_Subscribe_NilSubscriber(t *testing.T) {
	ctx := context.Background()
	adapter := &Adapter{
		subscriber: nil,
	}

	ch, err := adapter.Subscribe(ctx, "test.topic")

	assert.Nil(t, ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscriber is not initialized")
}

func TestAdapter_Flush_NilConnection(t *testing.T) {
	adapter := &Adapter{
		conn: nil,
	}

	err := adapter.Flush()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NATS connection is nil")
}

func TestAdapter_FlushTimeout_NilConnection(t *testing.T) {
	adapter := &Adapter{
		conn: nil,
	}

	err := adapter.FlushTimeout(5 * time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NATS connection is nil")
}

func TestAdapter_Close_NilComponents(t *testing.T) {
	adapter := &Adapter{
		publisher:  nil,
		subscriber: nil,
		conn:       nil,
	}

	err := adapter.Close()
	assert.NoError(t, err)
}

// Benchmark tests for performance validation
func BenchmarkConfigValidate(b *testing.B) {
	config := validConfig

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}

func BenchmarkConfigSetDefaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config := Config{}
		config.SetDefaults()
	}
}

func BenchmarkFormatValidationErrors(b *testing.B) {
	errors := map[string]error{
		"url":       assert.AnError,
		"client_id": assert.AnError,
		"timeout":   assert.AnError,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatValidationErrors(errors)
	}
}

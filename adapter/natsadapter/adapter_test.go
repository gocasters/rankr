package natsadapter

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements Logger interface for testing
type MockLogger struct {
	InfoCalls  []LogCall
	ErrorCalls []LogCall
}

type LogCall struct {
	Msg    string
	Fields []interface{}
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.InfoCalls = append(m.InfoCalls, LogCall{Msg: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.ErrorCalls = append(m.ErrorCalls, LogCall{Msg: msg, Fields: fields})
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		InfoCalls:  make([]LogCall, 0),
		ErrorCalls: make([]LogCall, 0),
	}
}

// TestServer helper
type TestServer struct {
	server *server.Server
	url    string
}

func startTestServer(t *testing.T) *TestServer {
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1, // Random port
		JetStream: true,
	}

	ns, err := server.NewServer(opts)
	require.NoError(t, err)

	go ns.Start()

	if !ns.ReadyForConnections(4 * time.Second) {
		t.Fatal("Unable to start NATS Server")
	}

	return &TestServer{
		server: ns,
		url:    ns.ClientURL(),
	}
}

func (ts *TestServer) Shutdown() {
	if ts.server != nil {
		ts.server.Shutdown()
		ts.server.WaitForShutdown()
	}
}

// Test Enum Conversions
func TestStorageTypeEnum_ToNatsStorage(t *testing.T) {
	tests := []struct {
		name     string
		storage  StorageTypeEnum
		expected nats.StorageType
	}{
		{
			name:     "memory storage",
			storage:  StorageMemory,
			expected: nats.MemoryStorage,
		},
		{
			name:     "file storage",
			storage:  StorageFile,
			expected: nats.FileStorage,
		},
		{
			name:     "invalid storage defaults to file",
			storage:  "invalid",
			expected: nats.FileStorage,
		},
		{
			name:     "empty storage defaults to file",
			storage:  "",
			expected: nats.FileStorage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.storage.ToNatsStorage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetentionPolicyEnum_ToNatsRetention(t *testing.T) {
	tests := []struct {
		name     string
		policy   RetentionPolicyEnum
		expected nats.RetentionPolicy
	}{
		{
			name:     "limits policy",
			policy:   RetentionLimits,
			expected: nats.LimitsPolicy,
		},
		{
			name:     "workqueue policy",
			policy:   RetentionWorkQueue,
			expected: nats.WorkQueuePolicy,
		},
		{
			name:     "interest policy",
			policy:   RetentionInterest,
			expected: nats.InterestPolicy,
		},
		{
			name:     "invalid policy defaults to limits",
			policy:   "invalid",
			expected: nats.LimitsPolicy,
		},
		{
			name:     "empty policy defaults to limits",
			policy:   "",
			expected: nats.LimitsPolicy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.ToNatsRetention()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test Config Validation
func TestConfig_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected Config
	}{
		{
			name: "all defaults applied",
			config: Config{
				URL:            "nats://localhost:4222",
				StreamName:     "TEST",
				StreamSubjects: []string{"test"},
			},
			expected: Config{
				URL:             "nats://localhost:4222",
				StreamName:      "TEST",
				StreamSubjects:  []string{"test"},
				MaxReconnects:   10,
				ReconnectWait:   2 * time.Second,
				Replicas:        1,
				StorageType:     StorageFile,
				RetentionPolicy: RetentionLimits,
			},
		},
		{
			name: "partial defaults",
			config: Config{
				URL:             "nats://localhost:4222",
				StreamName:      "TEST",
				StreamSubjects:  []string{"test"},
				MaxReconnects:   5,
				StorageType:     StorageMemory,
				RetentionPolicy: RetentionWorkQueue,
			},
			expected: Config{
				URL:             "nats://localhost:4222",
				StreamName:      "TEST",
				StreamSubjects:  []string{"test"},
				MaxReconnects:   5,
				ReconnectWait:   2 * time.Second,
				Replicas:        1,
				StorageType:     StorageMemory,
				RetentionPolicy: RetentionWorkQueue,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			config.SetDefaults()
			assert.Equal(t, tt.expected, config)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedErrs  int
		containsError string
	}{
		{
			name: "valid config",
			config: Config{
				URL:            "nats://localhost:4222",
				StreamName:     "TEST",
				StreamSubjects: []string{"test"},
				MaxMessages:    1000,
				MaxBytes:       1024,
			},
			expectedErrs: 0,
		},
		{
			name: "missing URL",
			config: Config{
				StreamName:     "TEST",
				StreamSubjects: []string{"test"},
			},
			expectedErrs:  1,
			containsError: "URL is required",
		},
		{
			name: "missing StreamName",
			config: Config{
				URL:            "nats://localhost:4222",
				StreamSubjects: []string{"test"},
			},
			expectedErrs:  1,
			containsError: "StreamName is required",
		},
		{
			name: "empty StreamSubjects",
			config: Config{
				URL:            "nats://localhost:4222",
				StreamName:     "TEST",
				StreamSubjects: []string{},
			},
			expectedErrs:  1,
			containsError: "StreamSubjects cannot be empty",
		},
		{
			name: "negative MaxMessages",
			config: Config{
				URL:            "nats://localhost:4222",
				StreamName:     "TEST",
				StreamSubjects: []string{"test"},
				MaxMessages:    -1,
			},
			expectedErrs:  1,
			containsError: "MaxMessages cannot be negative",
		},
		{
			name: "negative MaxBytes",
			config: Config{
				URL:            "nats://localhost:4222",
				StreamName:     "TEST",
				StreamSubjects: []string{"test"},
				MaxBytes:       -1,
			},
			expectedErrs:  1,
			containsError: "MaxBytes cannot be negative",
		},
		{
			name: "multiple errors",
			config: Config{
				MaxMessages: -1,
				MaxBytes:    -1,
			},
			expectedErrs: 5, // URL, StreamName, StreamSubjects, MaxMessages, MaxBytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()
			assert.Len(t, errs, tt.expectedErrs)

			if tt.containsError != "" {
				found := false
				for _, err := range errs {
					if err == tt.containsError {
						found = true
						break
					}
				}
				assert.True(t, found, "expected error '%s' not found", tt.containsError)
			}
		})
	}
}

// Test Adapter Creation
func TestNew_Success(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "TEST_STREAM",
		MaxReconnects:   3,
		ReconnectWait:   1 * time.Second,
		StreamSubjects:  []string{"test.subject"},
		MaxMessages:     1000,
		MaxBytes:        1024 * 1024,
		MaxAge:          1 * time.Hour,
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	defer adapter.Close()

	assert.NotNil(t, adapter.conn)
	assert.NotNil(t, adapter.js)
	assert.Equal(t, config.URL, adapter.config.URL)
	assert.Equal(t, config.StreamName, adapter.config.StreamName)

	// Verify logger was called for stream creation
	assert.NotEmpty(t, logger.InfoCalls)
	assert.Contains(t, logger.InfoCalls[0].Msg, "Created stream")
}

func TestNew_InvalidConfig(t *testing.T) {
	logger := NewMockLogger()
	config := Config{
		// Missing required fields
		StreamName: "TEST",
	}

	adapter, err := New(config, logger)
	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestNew_InvalidURL(t *testing.T) {
	logger := NewMockLogger()
	config := Config{
		URL:            "nats://invalid:4222",
		StreamName:     "TEST_STREAM",
		StreamSubjects: []string{"test"},
	}

	adapter, err := New(config, logger)
	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestNew_StreamCreationAndUpdate(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "UPDATE_STREAM",
		StreamSubjects:  []string{"update.subject"},
		MaxMessages:     1000,
		MaxBytes:        1024 * 1024,
		MaxAge:          1 * time.Hour,
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionLimits,
	}

	// First creation
	adapter1, err := New(config, logger)
	require.NoError(t, err)
	defer adapter1.Close()

	// Verify stream created
	info, err := adapter1.GetStreamInfo()
	require.NoError(t, err)
	assert.Equal(t, "UPDATE_STREAM", info.Config.Name)

	assert.Len(t, logger.InfoCalls, 1)
	assert.Contains(t, logger.InfoCalls[0].Msg, "Created stream")

	// Update stream
	config.MaxMessages = 2000
	adapter2, err := New(config, logger)
	require.NoError(t, err)
	defer adapter2.Close()

	// Verify stream updated (no new "Created stream" log)
	info2, err := adapter2.GetStreamInfo()
	require.NoError(t, err)
	assert.Equal(t, int64(2000), info2.Config.MaxMsgs)
	assert.Len(t, logger.InfoCalls, 1) // Still only 1 create log
}

// Test Publish
func TestAdapter_Publish_Success(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "PUBLISH_STREAM",
		StreamSubjects:  []string{"test.>"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	ctx := context.Background()
	data := []byte("test message")

	err = adapter.Publish(ctx, "test.subject", data)
	assert.NoError(t, err)

	// Verify message was published
	info, err := adapter.GetStreamInfo()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), info.State.Msgs)
}

func TestAdapter_Publish_InvalidSubject(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "PUBLISH_STREAM",
		StreamSubjects:  []string{"valid.subject"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	ctx := context.Background()
	data := []byte("test message")

	err = adapter.Publish(ctx, "invalid.subject", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish to subject")
}

func TestAdapter_Publish_ContextCancellation(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "PUBLISH_STREAM",
		StreamSubjects:  []string{"test.>"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := []byte("test message")
	err = adapter.Publish(ctx, "test.subject", data)
	assert.Error(t, err)
}

func TestAdapter_PublishAsync(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "ASYNC_STREAM",
		StreamSubjects:  []string{"async.>"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	data := []byte("async message")

	future, err := adapter.PublishAsync("async.test", data)
	require.NoError(t, err)
	require.NotNil(t, future)

	// Wait for ACK
	select {
	case <-future.Ok():
		// Success
	case err := <-future.Err():
		t.Fatalf("publish async failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for async publish")
	}

	// Verify message
	info, err := adapter.GetStreamInfo()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), info.State.Msgs)
}

// Test Pull Consumer
func TestAdapter_CreatePullConsumer_Success(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "CONSUMER_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	consumerConfig := PullConsumerConfig{
		Subject:     "events",
		DurableName: "test-consumer",
		BatchSize:   10,
		MaxWait:     1 * time.Second,
		MaxDeliver:  3,
		AckWait:     30 * time.Second,
	}

	consumer, err := adapter.CreatePullConsumer(consumerConfig)
	require.NoError(t, err)
	require.NotNil(t, consumer)
	defer consumer.Close()

	assert.NotNil(t, consumer.sub)
	assert.Equal(t, consumerConfig, consumer.config)
}

func TestPullConsumer_Fetch_NoMessages(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "FETCH_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	consumerConfig := PullConsumerConfig{
		Subject:     "events",
		DurableName: "test-consumer",
		BatchSize:   10,
		MaxWait:     100 * time.Millisecond,
		MaxDeliver:  3,
		AckWait:     30 * time.Second,
	}

	consumer, err := adapter.CreatePullConsumer(consumerConfig)
	require.NoError(t, err)
	defer consumer.Close()

	msgs, err := consumer.Fetch()
	assert.NoError(t, err)
	assert.Nil(t, msgs) // Should return nil when no messages
}

func TestPullConsumer_Fetch_WithMessages(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "FETCH_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	// Publish messages
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err = adapter.Publish(ctx, "events", []byte("message"))
		require.NoError(t, err)
	}

	consumerConfig := PullConsumerConfig{
		Subject:     "events",
		DurableName: "test-consumer",
		BatchSize:   3,
		MaxWait:     1 * time.Second,
		MaxDeliver:  3,
		AckWait:     30 * time.Second,
	}

	consumer, err := adapter.CreatePullConsumer(consumerConfig)
	require.NoError(t, err)
	defer consumer.Close()

	msgs, err := consumer.Fetch()
	require.NoError(t, err)
	assert.Len(t, msgs, 3) // Should fetch batch size

	// ACK messages
	for _, msg := range msgs {
		err = msg.Ack()
		assert.NoError(t, err)
	}

	// Fetch remaining
	msgs2, err := consumer.Fetch()
	require.NoError(t, err)
	assert.Len(t, msgs2, 2) // Remaining 2 messages
}

func TestPullConsumer_Fetch_NAK_Redelivery(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "NAK_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	// Publish message
	ctx := context.Background()
	err = adapter.Publish(ctx, "events", []byte("test"))
	require.NoError(t, err)

	consumerConfig := PullConsumerConfig{
		Subject:     "events",
		DurableName: "nak-consumer",
		BatchSize:   1,
		MaxWait:     1 * time.Second,
		MaxDeliver:  3,
		AckWait:     1 * time.Second,
	}

	consumer, err := adapter.CreatePullConsumer(consumerConfig)
	require.NoError(t, err)
	defer consumer.Close()

	// Fetch and NAK
	msgs, err := consumer.Fetch()
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	err = msgs[0].Nak()
	assert.NoError(t, err)

	// Should be redelivered
	time.Sleep(1500 * time.Millisecond) // Wait for redelivery
	msgs2, err := consumer.Fetch()
	require.NoError(t, err)
	assert.Len(t, msgs2, 1) // Message redelivered
}

func TestPullConsumer_GetConsumerInfo(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "INFO_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	consumerConfig := PullConsumerConfig{
		Subject:     "events",
		DurableName: "info-consumer",
		BatchSize:   10,
		MaxWait:     1 * time.Second,
		MaxDeliver:  3,
		AckWait:     30 * time.Second,
	}

	consumer, err := adapter.CreatePullConsumer(consumerConfig)
	require.NoError(t, err)
	defer consumer.Close()

	info, err := consumer.GetConsumerInfo()
	require.NoError(t, err)
	assert.Equal(t, "info-consumer", info.Name)
	assert.Equal(t, "INFO_STREAM", info.Stream)
	assert.Equal(t, 3, info.Config.MaxDeliver)
}

// Test Close
func TestAdapter_Close(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "CLOSE_STREAM",
		StreamSubjects:  []string{"test"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)

	err = adapter.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	assert.False(t, adapter.conn.IsConnected())
}

func TestPullConsumer_Close(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:             ts.url,
		StreamName:      "CLOSE_CONSUMER_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	consumerConfig := PullConsumerConfig{
		Subject:     "events",
		DurableName: "close-test",
		BatchSize:   10,
		MaxWait:     1 * time.Second,
		MaxDeliver:  3,
		AckWait:     30 * time.Second,
	}

	consumer, err := adapter.CreatePullConsumer(consumerConfig)
	require.NoError(t, err)

	err = consumer.Close()
	assert.NoError(t, err)
}

// Test Logger Integration
func TestAdapter_LoggerIntegration(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	logger := NewMockLogger()
	config := Config{
		URL:            ts.url,
		StreamName:     "LOGGER_STREAM",
		StreamSubjects: []string{"test"},
		StorageType:    StorageMemory,
	}

	adapter, err := New(config, logger)
	require.NoError(t, err)
	defer adapter.Close()

	// Verify stream creation was logged
	assert.NotEmpty(t, logger.InfoCalls)
	assert.Contains(t, logger.InfoCalls[0].Msg, "Created stream")

	// Simulate disconnect/reconnect (would need to force disconnect in real scenario)
	// For now, just verify the handlers are set up
	assert.NotNil(t, adapter.conn)
}

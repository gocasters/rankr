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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.ToNatsRetention()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test Adapter Creation
func TestNew_Success(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

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

	adapter, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	defer adapter.Close()

	assert.NotNil(t, adapter.conn)
	assert.NotNil(t, adapter.js)
	assert.Equal(t, config, adapter.config)
}

func TestNew_InvalidURL(t *testing.T) {
	config := Config{
		URL:        "nats://invalid:4222",
		StreamName: "TEST_STREAM",
	}

	adapter, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "connection closed")
}

func TestNew_StreamCreation(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "NEW_STREAM",
		StreamSubjects:  []string{"new.subject"},
		MaxMessages:     1000,
		MaxBytes:        1024 * 1024,
		MaxAge:          1 * time.Hour,
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionLimits,
	}

	adapter, err := New(config)
	require.NoError(t, err)
	defer adapter.Close()

	// Verify stream exists
	info, err := adapter.GetStreamInfo()
	require.NoError(t, err)
	assert.Equal(t, "NEW_STREAM", info.Config.Name)
	assert.Equal(t, []string{"new.subject"}, info.Config.Subjects)
}

// Test Publish
func TestAdapter_Publish(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "PUBLISH_STREAM",
		StreamSubjects:  []string{"test.>"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
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

	config := Config{
		URL:             ts.url,
		StreamName:      "PUBLISH_STREAM",
		StreamSubjects:  []string{"valid.subject"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
	require.NoError(t, err)
	defer adapter.Close()

	ctx := context.Background()
	data := []byte("test message")

	err = adapter.Publish(ctx, "invalid.subject", data)
	assert.Error(t, err)
}

func TestAdapter_PublishAsync(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "ASYNC_STREAM",
		StreamSubjects:  []string{"async.>"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
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
}

// Test Pull Consumer
func TestAdapter_CreatePullConsumer(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "CONSUMER_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
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

	config := Config{
		URL:             ts.url,
		StreamName:      "FETCH_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
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

	config := Config{
		URL:             ts.url,
		StreamName:      "FETCH_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
	require.NoError(t, err)
	defer adapter.Close()

	// Publish some messages
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
}

func TestPullConsumer_GetConsumerInfo(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "INFO_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
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
	defer consumer.Close()

	info, err := consumer.GetConsumerInfo()
	require.NoError(t, err)
	assert.Equal(t, "test-consumer", info.Name)
	assert.Equal(t, "INFO_STREAM", info.Stream)
}

// Test Close
func TestAdapter_Close(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "CLOSE_STREAM",
		StreamSubjects:  []string{"test"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
	require.NoError(t, err)

	err = adapter.Close()
	assert.NoError(t, err)

	// Verify connection is closed
	assert.False(t, adapter.conn.IsConnected())
}

func TestPullConsumer_Close(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Shutdown()

	config := Config{
		URL:             ts.url,
		StreamName:      "CLOSE_CONSUMER_STREAM",
		StreamSubjects:  []string{"events"},
		StorageType:     StorageMemory,
		Replicas:        1,
		RetentionPolicy: RetentionWorkQueue,
	}

	adapter, err := New(config)
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

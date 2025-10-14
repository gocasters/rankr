package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/pkg/realtimeconstant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockMessageSubscriber struct {
	messages      chan *message.Message
	subscriptions map[string]chan *message.Message
	closed        bool
	mu            sync.RWMutex
}

func NewMockMessageSubscriber() *MockMessageSubscriber {
	return &MockMessageSubscriber{
		messages:      make(chan *message.Message, 100),
		subscriptions: make(map[string]chan *message.Message),
	}
}

func (m *MockMessageSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan *message.Message, 100)
	m.subscriptions[topic] = ch
	return ch, nil
}

func (m *MockMessageSubscriber) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	for _, ch := range m.subscriptions {
		close(ch)
	}
	return nil
}

func (m *MockMessageSubscriber) PublishToTopic(topic string, msg *message.Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ch, ok := m.subscriptions[topic]; ok {
		select {
		case ch <- msg:
		default:

		}
	}
}

func (m *MockMessageSubscriber) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

type MockRealtimeService struct {
	broadcastEvents []BroadcastEventRequest
	mu              sync.RWMutex
}

func NewMockRealtimeService() *MockRealtimeService {
	return &MockRealtimeService{
		broadcastEvents: make([]BroadcastEventRequest, 0),
	}
}

func (m *MockRealtimeService) RegisterClient(client *Client) {}

func (m *MockRealtimeService) UnregisterClient(clientID string) {}

func (m *MockRealtimeService) SubscribeTopics(ctx context.Context, clientID string, req SubscribeRequest) (SubscribeResponse, error) {
	return SubscribeResponse{Success: true}, nil
}

func (m *MockRealtimeService) UnsubscribeTopics(ctx context.Context, clientID string, req UnsubscribeRequest) (UnsubscribeResponse, error) {
	return UnsubscribeResponse{Success: true}, nil
}

func (m *MockRealtimeService) BroadcastEvent(ctx context.Context, req BroadcastEventRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.broadcastEvents = append(m.broadcastEvents, req)
	return nil
}

func (m *MockRealtimeService) GetConnectedClientCount() int {
	return 0
}

func (m *MockRealtimeService) GetBroadcastEvents() []BroadcastEventRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.broadcastEvents
}

func setupTestSubscriber() (*Subscriber, *MockMessageSubscriber, *MockRealtimeService) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockSubscriber := NewMockMessageSubscriber()
	mockService := NewMockRealtimeService()
	topics := []string{
		realtimeconstant.TopicTaskCreated,
		realtimeconstant.TopicTaskUpdated,
		realtimeconstant.TopicContributorCreated,
	}

	subscriber := NewSubscriber(mockSubscriber, mockService, topics, logger)
	return subscriber, mockSubscriber, mockService
}

func TestNewSubscriber(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockSubscriber := NewMockMessageSubscriber()
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, logger)
	topics := []string{realtimeconstant.TopicTaskCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, logger)

	assert.NotNil(t, subscriber)
	assert.NotNil(t, subscriber.Subscriber)
	assert.NotNil(t, subscriber.Service)
	assert.NotNil(t, subscriber.Logger)
	assert.Equal(t, topics, subscriber.Topics)
}

func TestSubscriber_Start(t *testing.T) {
	subscriber, mockSubscriber, _ := setupTestSubscriber()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	assert.Equal(t, 3, len(mockSubscriber.subscriptions))
	assert.Contains(t, mockSubscriber.subscriptions, realtimeconstant.TopicTaskCreated)
	assert.Contains(t, mockSubscriber.subscriptions, realtimeconstant.TopicTaskUpdated)
	assert.Contains(t, mockSubscriber.subscriptions, realtimeconstant.TopicContributorCreated)
}

func TestSubscriber_ProcessMessages(t *testing.T) {
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mockSubscriber := NewMockMessageSubscriber()
	topics := []string{realtimeconstant.TopicTaskCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	client := &Client{
		ID:            "test-client",
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
	client.Subscriptions[realtimeconstant.TopicTaskCreated] = true
	service.RegisterClient(client)

	payload := map[string]interface{}{
		"task_id":   "123",
		"task_name": "Test Task",
		"status":    "created",
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := message.NewMessage(uuid.New().String(), payloadBytes)

	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg)

	time.Sleep(200 * time.Millisecond)

	select {
	case eventData := <-client.Send:
		var event Event
		err := json.Unmarshal(eventData, &event)
		require.NoError(t, err)
		assert.Equal(t, realtimeconstant.TopicTaskCreated, event.Topic)
		assert.Equal(t, "123", event.Payload["task_id"])
		assert.Equal(t, "Test Task", event.Payload["task_name"])
	default:
		t.Fatal("No event received by client")
	}
}

func TestSubscriber_ProcessMultipleMessages(t *testing.T) {
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mockSubscriber := NewMockMessageSubscriber()
	topics := []string{realtimeconstant.TopicTaskCreated, realtimeconstant.TopicTaskUpdated, realtimeconstant.TopicContributorCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	client1 := &Client{
		ID:            "client-1",
		Send:          make(chan []byte, 256),
		Subscriptions: map[string]bool{realtimeconstant.TopicTaskCreated: true, realtimeconstant.TopicTaskUpdated: true, realtimeconstant.TopicContributorCreated: true},
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
	service.RegisterClient(client1)

	messages := []struct {
		topic   string
		payload map[string]interface{}
	}{
		{
			topic: realtimeconstant.TopicTaskCreated,
			payload: map[string]interface{}{
				"task_id": "1",
				"action":  "created",
			},
		},
		{
			topic: realtimeconstant.TopicTaskUpdated,
			payload: map[string]interface{}{
				"task_id": "2",
				"action":  "updated",
			},
		},
		{
			topic: realtimeconstant.TopicContributorCreated,
			payload: map[string]interface{}{
				"contributor_id": "3",
				"action":         "created",
			},
		},
	}

	for _, m := range messages {
		payloadBytes, _ := json.Marshal(m.payload)
		msg := message.NewMessage(uuid.New().String(), payloadBytes)
		mockSubscriber.PublishToTopic(m.topic, msg)
	}

	time.Sleep(300 * time.Millisecond)

	receivedTopics := make(map[string]int)
	for i := 0; i < 3; i++ {
		select {
		case eventData := <-client1.Send:
			var event Event
			err := json.Unmarshal(eventData, &event)
			require.NoError(t, err)
			receivedTopics[event.Topic]++
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Only received %d out of 3 expected events", i)
		}
	}

	assert.Equal(t, 1, receivedTopics[realtimeconstant.TopicTaskCreated])
	assert.Equal(t, 1, receivedTopics[realtimeconstant.TopicTaskUpdated])
	assert.Equal(t, 1, receivedTopics[realtimeconstant.TopicContributorCreated])
}

func TestSubscriber_HandleInvalidJSON(t *testing.T) {
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mockSubscriber := NewMockMessageSubscriber()
	topics := []string{realtimeconstant.TopicTaskCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	client := &Client{
		ID:            "test-client",
		Send:          make(chan []byte, 256),
		Subscriptions: map[string]bool{realtimeconstant.TopicTaskCreated: true},
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
	service.RegisterClient(client)

	invalidJSON := []byte("{invalid json}")
	msg := message.NewMessage(uuid.New().String(), invalidJSON)

	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg)

	time.Sleep(200 * time.Millisecond)

	select {
	case <-client.Send:
		t.Fatal("Client should not have received an event for invalid JSON")
	default:

	}
}

func TestSubscriber_ContextCancellation(t *testing.T) {
	subscriber, mockSubscriber, mockService := setupTestSubscriber()
	ctx, cancel := context.WithCancel(context.Background())

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	payload := map[string]interface{}{"test": "data"}
	payloadBytes, _ := json.Marshal(payload)
	msg := message.NewMessage(uuid.New().String(), payloadBytes)
	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg)

	time.Sleep(100 * time.Millisecond)
	events := mockService.GetBroadcastEvents()
	initialCount := len(events)

	cancel()

	time.Sleep(100 * time.Millisecond)

	msg2 := message.NewMessage(uuid.New().String(), payloadBytes)
	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg2)

	time.Sleep(100 * time.Millisecond)

	events = mockService.GetBroadcastEvents()
	assert.Equal(t, initialCount, len(events))
}

func TestSubscriber_Stop(t *testing.T) {
	subscriber, mockSubscriber, _ := setupTestSubscriber()
	ctx := context.Background()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	err = subscriber.Stop()
	require.NoError(t, err)

	assert.True(t, mockSubscriber.IsClosed())
}

func TestSubscriber_MessageAcknowledgement(t *testing.T) {
	subscriber, mockSubscriber, _ := setupTestSubscriber()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	payload := map[string]interface{}{"test": "data"}
	payloadBytes, _ := json.Marshal(payload)
	msg := message.NewMessage(uuid.New().String(), payloadBytes)

	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg)

	time.Sleep(100 * time.Millisecond)

}

func TestSubscriber_ConcurrentMessageProcessing(t *testing.T) {
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mockSubscriber := NewMockMessageSubscriber()
	topics := []string{realtimeconstant.TopicTaskCreated, realtimeconstant.TopicTaskUpdated, realtimeconstant.TopicContributorCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	client := &Client{
		ID:   "test-client",
		Send: make(chan []byte, 256),
		Subscriptions: map[string]bool{
			realtimeconstant.TopicTaskCreated:        true,
			realtimeconstant.TopicTaskUpdated:        true,
			realtimeconstant.TopicContributorCreated: true,
		},
		ConnectedAt:  time.Now(),
		LastActiveAt: time.Now(),
	}
	service.RegisterClient(client)

	numMessages := 50
	var wg sync.WaitGroup
	wg.Add(numMessages)

	for i := 0; i < numMessages; i++ {
		go func(id int) {
			defer wg.Done()

			payload := map[string]interface{}{
				"id":   id,
				"data": "test",
			}
			payloadBytes, _ := json.Marshal(payload)
			msg := message.NewMessage(uuid.New().String(), payloadBytes)

			topic := realtimeconstant.TopicTaskCreated
			if id%3 == 0 {
				topic = realtimeconstant.TopicTaskUpdated
			} else if id%3 == 1 {
				topic = realtimeconstant.TopicContributorCreated
			}

			mockSubscriber.PublishToTopic(topic, msg)
		}(i)
	}

	wg.Wait()

	time.Sleep(1 * time.Second)

	receivedCount := 0
	for {
		select {
		case <-client.Send:
			receivedCount++
		default:

			assert.Equal(t, numMessages, receivedCount, "Should have received all %d messages", numMessages)
			return
		}
	}
}

func TestSubscriber_EmptyPayload(t *testing.T) {
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mockSubscriber := NewMockMessageSubscriber()
	topics := []string{realtimeconstant.TopicTaskCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	client := &Client{
		ID:            "test-client",
		Send:          make(chan []byte, 256),
		Subscriptions: map[string]bool{realtimeconstant.TopicTaskCreated: true},
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
	service.RegisterClient(client)

	emptyPayload := []byte("{}")
	msg := message.NewMessage(uuid.New().String(), emptyPayload)

	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg)

	time.Sleep(200 * time.Millisecond)

	select {
	case eventData := <-client.Send:
		var event Event
		err := json.Unmarshal(eventData, &event)
		require.NoError(t, err)
		assert.Equal(t, realtimeconstant.TopicTaskCreated, event.Topic)
		assert.NotNil(t, event.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No event received")
	}
}

func TestSubscriber_ComplexPayload(t *testing.T) {
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mockSubscriber := NewMockMessageSubscriber()
	topics := []string{realtimeconstant.TopicTaskCreated}

	subscriber := NewSubscriber(mockSubscriber, service, topics, slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	client := &Client{
		ID:            "test-client",
		Send:          make(chan []byte, 256),
		Subscriptions: map[string]bool{realtimeconstant.TopicTaskCreated: true},
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
	service.RegisterClient(client)

	complexPayload := map[string]interface{}{
		"task_id": "123",
		"task": map[string]interface{}{
			"name":        "Complex Task",
			"description": "A task with nested data",
			"metadata": map[string]interface{}{
				"priority": "high",
				"tags":     []string{"urgent", "backend"},
			},
		},
		"contributor": map[string]interface{}{
			"id":       "456",
			"username": "testuser",
		},
		"timestamp": time.Now().Unix(),
	}

	payloadBytes, _ := json.Marshal(complexPayload)
	msg := message.NewMessage(uuid.New().String(), payloadBytes)

	mockSubscriber.PublishToTopic(realtimeconstant.TopicTaskCreated, msg)

	time.Sleep(200 * time.Millisecond)

	select {
	case eventData := <-client.Send:
		var event Event
		err := json.Unmarshal(eventData, &event)
		require.NoError(t, err)
		assert.Equal(t, realtimeconstant.TopicTaskCreated, event.Topic)

		taskData, ok := event.Payload["task"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Complex Task", taskData["name"])

		metadata, ok := taskData["metadata"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "high", metadata["priority"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No event received")
	}
}

func TestSubscriber_NoTopics(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockSubscriber := NewMockMessageSubscriber()
	mockConnectionStore := NewMockConnectionStore()
	service := NewService(mockConnectionStore, logger)
	topics := []string{}

	subscriber := NewSubscriber(mockSubscriber, service, topics, logger)
	ctx := context.Background()

	err := subscriber.Start(ctx)
	require.NoError(t, err)

	assert.Equal(t, 0, len(mockSubscriber.subscriptions))
}

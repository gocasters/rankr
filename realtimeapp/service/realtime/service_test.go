package realtime

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/realtimeconstant"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockConnectionStore struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewMockConnectionStore() *MockConnectionStore {
	return &MockConnectionStore{
		clients: make(map[string]*Client),
	}
}

func (m *MockConnectionStore) AddClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client
}

func (m *MockConnectionStore) RemoveClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, clientID)
}

func (m *MockConnectionStore) GetClient(clientID string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[clientID]
	return client, ok
}

func (m *MockConnectionStore) GetAllClients() []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients
}

func (m *MockConnectionStore) GetClientsByTopic(topic string) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]*Client, 0)
	for _, client := range m.clients {
		if client.Subscriptions[topic] {
			clients = append(clients, client)
		}
	}
	return clients
}

func setupTestService() (Service, *MockConnectionStore) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := NewMockConnectionStore()
	service := NewService(store, logger)
	return service, store
}

func createTestClient(id string) *Client {
	return &Client{
		ID:            id,
		Conn:          &websocket.Conn{},
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
}

func TestNewService(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := NewMockConnectionStore()

	service := NewService(store, logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.ConnectionStore)
	assert.NotNil(t, service.Logger)
}

func TestService_RegisterClient(t *testing.T) {
	service, store := setupTestService()
	client := createTestClient("client-1")

	service.RegisterClient(client)

	retrieved, ok := store.GetClient("client-1")
	assert.True(t, ok)
	assert.Equal(t, client, retrieved)
}

func TestService_RegisterMultipleClients(t *testing.T) {
	service, store := setupTestService()

	for i := 1; i <= 5; i++ {
		client := createTestClient(string(rune('0' + i)))
		service.RegisterClient(client)
	}

	clients := store.GetAllClients()
	assert.Equal(t, 5, len(clients))
}

func TestService_UnregisterClient(t *testing.T) {
	service, store := setupTestService()
	client := createTestClient("client-1")

	service.RegisterClient(client)
	assert.Equal(t, 1, len(store.GetAllClients()))

	service.UnregisterClient("client-1")

	_, ok := store.GetClient("client-1")
	assert.False(t, ok)
	assert.Equal(t, 0, len(store.GetAllClients()))
}

func TestService_UnregisterNonExistentClient(t *testing.T) {
	service, store := setupTestService()

	service.UnregisterClient("non-existent")
	assert.Equal(t, 0, len(store.GetAllClients()))
}

func TestService_SubscribeTopics(t *testing.T) {
	service, store := setupTestService()
	client := createTestClient("client-1")
	service.RegisterClient(client)
	ctx := context.Background()

	t.Run("should subscribe to single topic", func(t *testing.T) {
		req := SubscribeRequest{
			Topics: []string{"task.created"},
		}

		resp, err := service.SubscribeTopics(ctx, "client-1", req)

		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, []string{"task.created"}, resp.Topics)

		retrieved, _ := store.GetClient("client-1")
		assert.True(t, retrieved.Subscriptions["task.created"])
	})

	t.Run("should subscribe to multiple topics", func(t *testing.T) {
		req := SubscribeRequest{
			Topics: []string{"task.updated", "task.completed", "contributor.created"},
		}

		resp, err := service.SubscribeTopics(ctx, "client-1", req)

		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, 3, len(resp.Topics))

		retrieved, _ := store.GetClient("client-1")
		assert.True(t, retrieved.Subscriptions["task.updated"])
		assert.True(t, retrieved.Subscriptions["task.completed"])
		assert.True(t, retrieved.Subscriptions["contributor.created"])
	})

	t.Run("should return error for non-existent client", func(t *testing.T) {
		req := SubscribeRequest{
			Topics: []string{"task.created"},
		}

		resp, err := service.SubscribeTopics(ctx, "non-existent", req)

		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "client not found", resp.Message)
	})
}

func TestService_UnsubscribeTopics(t *testing.T) {
	service, store := setupTestService()
	client := createTestClient("client-1")
	client.Subscriptions["task.created"] = true
	client.Subscriptions["task.updated"] = true
	client.Subscriptions["contributor.created"] = true
	service.RegisterClient(client)
	ctx := context.Background()

	t.Run("should unsubscribe from single topic", func(t *testing.T) {
		req := UnsubscribeRequest{
			Topics: []string{"task.created"},
		}

		resp, err := service.UnsubscribeTopics(ctx, "client-1", req)

		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, []string{"task.created"}, resp.Topics)

		retrieved, _ := store.GetClient("client-1")
		assert.False(t, retrieved.Subscriptions["task.created"])
		assert.True(t, retrieved.Subscriptions["task.updated"])
		assert.True(t, retrieved.Subscriptions["contributor.created"])
	})

	t.Run("should unsubscribe from multiple topics", func(t *testing.T) {
		req := UnsubscribeRequest{
			Topics: []string{"task.updated", "contributor.created"},
		}

		resp, err := service.UnsubscribeTopics(ctx, "client-1", req)

		require.NoError(t, err)
		assert.True(t, resp.Success)

		retrieved, _ := store.GetClient("client-1")
		assert.False(t, retrieved.Subscriptions["task.updated"])
		assert.False(t, retrieved.Subscriptions["contributor.created"])
	})

	t.Run("should return error for non-existent client", func(t *testing.T) {
		req := UnsubscribeRequest{
			Topics: []string{"task.created"},
		}

		resp, err := service.UnsubscribeTopics(ctx, "non-existent", req)

		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "client not found", resp.Message)
	})

	t.Run("should handle unsubscribing from non-subscribed topic", func(t *testing.T) {
		req := UnsubscribeRequest{
			Topics: []string{"non.subscribed.topic"},
		}

		resp, err := service.UnsubscribeTopics(ctx, "client-1", req)

		require.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

func TestService_BroadcastEvent(t *testing.T) {
	service, store := setupTestService()
	ctx := context.Background()

	client1 := createTestClient("client-1")
	client1.Subscriptions[realtimeconstant.TopicTaskCreated] = true
	service.RegisterClient(client1)

	client2 := createTestClient("client-2")
	client2.Subscriptions[realtimeconstant.TopicTaskCreated] = true
	service.RegisterClient(client2)

	client3 := createTestClient("client-3")
	client3.Subscriptions[realtimeconstant.TopicContributorCreated] = true
	service.RegisterClient(client3)

	t.Run("should broadcast event to subscribed clients", func(t *testing.T) {
		req := BroadcastEventRequest{
			Topic: realtimeconstant.TopicTaskCreated,
			Payload: map[string]interface{}{
				"task_id":   "123",
				"task_name": "Test Task",
			},
		}

		err := service.BroadcastEvent(ctx, req)
		require.NoError(t, err)

		select {
		case msg := <-client1.Send:
			assert.NotNil(t, msg)
			assert.Contains(t, string(msg), realtimeconstant.TopicTaskCreated)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("client1 did not receive event")
		}

		select {
		case msg := <-client2.Send:
			assert.NotNil(t, msg)
			assert.Contains(t, string(msg), realtimeconstant.TopicTaskCreated)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("client2 did not receive event")
		}

		select {
		case <-client3.Send:
			t.Fatal("client3 should not have received the event")
		case <-time.After(50 * time.Millisecond):

		}
	})

	t.Run("should broadcast to different topic", func(t *testing.T) {
		req := BroadcastEventRequest{
			Topic: realtimeconstant.TopicContributorCreated,
			Payload: map[string]interface{}{
				"contributor_id": "456",
				"username":       "testuser",
			},
		}

		err := service.BroadcastEvent(ctx, req)
		require.NoError(t, err)

		select {
		case msg := <-client3.Send:
			assert.NotNil(t, msg)
			assert.Contains(t, string(msg), realtimeconstant.TopicContributorCreated)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("client3 did not receive event")
		}
	})

	t.Run("should handle broadcast to topic with no subscribers", func(t *testing.T) {
		req := BroadcastEventRequest{
			Topic: "no.subscribers.topic",
			Payload: map[string]interface{}{
				"test": "data",
			},
		}

		err := service.BroadcastEvent(ctx, req)
		require.NoError(t, err)

		clients := store.GetClientsByTopic("no.subscribers.topic")
		assert.Equal(t, 0, len(clients))
	})
}

func TestService_GetConnectedClientCount(t *testing.T) {
	service, _ := setupTestService()

	t.Run("should return zero when no clients", func(t *testing.T) {
		count := service.GetConnectedClientCount()
		assert.Equal(t, 0, count)
	})

	t.Run("should return correct count", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			client := createTestClient(string(rune('0' + i)))
			service.RegisterClient(client)
		}

		count := service.GetConnectedClientCount()
		assert.Equal(t, 3, count)
	})

	t.Run("should update count after unregister", func(t *testing.T) {
		service.UnregisterClient("1")
		count := service.GetConnectedClientCount()
		assert.Equal(t, 2, count)
	})
}

func TestService_ConcurrentOperations(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	var wg sync.WaitGroup

	numClients := 50
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			client := createTestClient(string(rune('0' + id)))
			service.RegisterClient(client)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, numClients, service.GetConnectedClientCount())

	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			req := SubscribeRequest{
				Topics: []string{realtimeconstant.TopicTaskCreated, realtimeconstant.TopicTaskUpdated},
			}
			service.SubscribeTopics(ctx, string(rune('0'+id)), req)
		}(i)
	}
	wg.Wait()

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			req := BroadcastEventRequest{
				Topic: realtimeconstant.TopicTaskCreated,
				Payload: map[string]interface{}{
					"test": "data",
				},
			}
			service.BroadcastEvent(ctx, req)
		}()
	}
	wg.Wait()

	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			service.UnregisterClient(string(rune('0' + id)))
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 0, service.GetConnectedClientCount())
}

func TestService_BroadcastEvent_FullChannelHandling(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	client := createTestClient("client-1")
	client.Send = make(chan []byte, 2)
	client.Subscriptions[realtimeconstant.TopicTaskCreated] = true
	service.RegisterClient(client)

	client.Send <- []byte("message1")
	client.Send <- []byte("message2")

	req := BroadcastEventRequest{
		Topic: realtimeconstant.TopicTaskCreated,
		Payload: map[string]interface{}{
			"test": "data",
		},
	}

	err := service.BroadcastEvent(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, 2, len(client.Send))
}

func TestService_SubscriptionPersistence(t *testing.T) {
	service, store := setupTestService()
	ctx := context.Background()

	client := createTestClient("client-1")
	service.RegisterClient(client)

	req := SubscribeRequest{
		Topics: []string{
			realtimeconstant.TopicTaskCreated,
			realtimeconstant.TopicTaskUpdated,
			realtimeconstant.TopicContributorCreated,
		},
	}

	_, err := service.SubscribeTopics(ctx, "client-1", req)
	require.NoError(t, err)

	retrieved, ok := store.GetClient("client-1")
	require.True(t, ok)
	assert.True(t, retrieved.Subscriptions[realtimeconstant.TopicTaskCreated])
	assert.True(t, retrieved.Subscriptions[realtimeconstant.TopicTaskUpdated])
	assert.True(t, retrieved.Subscriptions[realtimeconstant.TopicContributorCreated])

	unsub := UnsubscribeRequest{
		Topics: []string{realtimeconstant.TopicTaskUpdated},
	}
	_, err = service.UnsubscribeTopics(ctx, "client-1", unsub)
	require.NoError(t, err)

	retrieved, ok = store.GetClient("client-1")
	require.True(t, ok)
	assert.True(t, retrieved.Subscriptions[realtimeconstant.TopicTaskCreated])
	assert.False(t, retrieved.Subscriptions[realtimeconstant.TopicTaskUpdated])
	assert.True(t, retrieved.Subscriptions[realtimeconstant.TopicContributorCreated])
}

package repository

import (
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gocasters/rankr/realtimeapp/service/realtime"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConnectionStore() *ConnectionStore {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return NewConnectionStore(logger)
}

func createMockClient(id string) *realtime.Client {
	return &realtime.Client{
		ID:            id,
		Conn:          &websocket.Conn{},
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}
}

func TestNewConnectionStore(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := NewConnectionStore(logger)

	assert.NotNil(t, store)
	assert.NotNil(t, store.Clients)
	assert.NotNil(t, store.Logger)
	assert.Equal(t, 0, len(store.Clients))
}

func TestConnectionStore_AddClient(t *testing.T) {
	store := setupTestConnectionStore()
	client := createMockClient("client-1")

	store.AddClient(client)

	assert.Equal(t, 1, len(store.Clients))
	assert.Equal(t, client, store.Clients["client-1"])
}

func TestConnectionStore_AddMultipleClients(t *testing.T) {
	store := setupTestConnectionStore()

	for i := 1; i <= 5; i++ {
		client := createMockClient(string(rune('0' + i)))
		store.AddClient(client)
	}

	assert.Equal(t, 5, len(store.Clients))
}

func TestConnectionStore_RemoveClient(t *testing.T) {
	store := setupTestConnectionStore()
	client := createMockClient("client-1")

	store.AddClient(client)
	assert.Equal(t, 1, len(store.Clients))

	store.RemoveClient("client-1")
	assert.Equal(t, 0, len(store.Clients))
}

func TestConnectionStore_RemoveNonExistentClient(t *testing.T) {
	store := setupTestConnectionStore()

	store.RemoveClient("non-existent")
	assert.Equal(t, 0, len(store.Clients))
}

func TestConnectionStore_GetClient(t *testing.T) {
	store := setupTestConnectionStore()
	client := createMockClient("client-1")

	store.AddClient(client)

	t.Run("should get existing client", func(t *testing.T) {
		retrieved, ok := store.GetClient("client-1")
		assert.True(t, ok)
		assert.Equal(t, client, retrieved)
	})

	t.Run("should not get non-existent client", func(t *testing.T) {
		retrieved, ok := store.GetClient("non-existent")
		assert.False(t, ok)
		assert.Nil(t, retrieved)
	})
}

func TestConnectionStore_GetAllClients(t *testing.T) {
	store := setupTestConnectionStore()

	t.Run("should return empty slice when no clients", func(t *testing.T) {
		clients := store.GetAllClients()
		assert.NotNil(t, clients)
		assert.Equal(t, 0, len(clients))
	})

	t.Run("should return all clients", func(t *testing.T) {
		client1 := createMockClient("client-1")
		client2 := createMockClient("client-2")
		client3 := createMockClient("client-3")

		store.AddClient(client1)
		store.AddClient(client2)
		store.AddClient(client3)

		clients := store.GetAllClients()
		assert.Equal(t, 3, len(clients))

		clientMap := make(map[string]*realtime.Client)
		for _, c := range clients {
			clientMap[c.ID] = c
		}

		assert.Contains(t, clientMap, "client-1")
		assert.Contains(t, clientMap, "client-2")
		assert.Contains(t, clientMap, "client-3")
	})
}

func TestConnectionStore_GetClientsByTopic(t *testing.T) {
	store := setupTestConnectionStore()

	client1 := createMockClient("client-1")
	client1.Subscriptions["task.created"] = true
	client1.Subscriptions["task.updated"] = true

	client2 := createMockClient("client-2")
	client2.Subscriptions["task.created"] = true

	client3 := createMockClient("client-3")
	client3.Subscriptions["contributor.created"] = true

	store.AddClient(client1)
	store.AddClient(client2)
	store.AddClient(client3)

	t.Run("should get clients subscribed to task.created", func(t *testing.T) {
		clients := store.GetClientsByTopic("task.created")
		assert.Equal(t, 2, len(clients))

		clientIDs := make([]string, len(clients))
		for i, c := range clients {
			clientIDs[i] = c.ID
		}

		assert.Contains(t, clientIDs, "client-1")
		assert.Contains(t, clientIDs, "client-2")
	})

	t.Run("should get clients subscribed to task.updated", func(t *testing.T) {
		clients := store.GetClientsByTopic("task.updated")
		assert.Equal(t, 1, len(clients))
		assert.Equal(t, "client-1", clients[0].ID)
	})

	t.Run("should get clients subscribed to contributor.created", func(t *testing.T) {
		clients := store.GetClientsByTopic("contributor.created")
		assert.Equal(t, 1, len(clients))
		assert.Equal(t, "client-3", clients[0].ID)
	})

	t.Run("should return empty for topic with no subscribers", func(t *testing.T) {
		clients := store.GetClientsByTopic("non.existent.topic")
		assert.Equal(t, 0, len(clients))
	})
}

func TestConnectionStore_ClientCount(t *testing.T) {
	store := setupTestConnectionStore()

	assert.Equal(t, 0, store.ClientCount())

	store.AddClient(createMockClient("client-1"))
	assert.Equal(t, 1, store.ClientCount())

	store.AddClient(createMockClient("client-2"))
	store.AddClient(createMockClient("client-3"))
	assert.Equal(t, 3, store.ClientCount())

	store.RemoveClient("client-2")
	assert.Equal(t, 2, store.ClientCount())
}

func TestConnectionStore_ConcurrentAccess(t *testing.T) {
	store := setupTestConnectionStore()
	var wg sync.WaitGroup

	numClients := 100
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			client := createMockClient(string(rune('0' + id)))
			store.AddClient(client)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, numClients, store.ClientCount())

	wg.Add(numClients * 2)
	for i := 0; i < numClients; i++ {

		go func(id int) {
			defer wg.Done()
			store.GetClient(string(rune('0' + id)))
			store.GetAllClients()
		}(i)

		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				store.RemoveClient(string(rune('0' + id)))
			}
		}(i)
	}
	wg.Wait()

	count := store.ClientCount()
	assert.Greater(t, count, 0)
	assert.LessOrEqual(t, count, numClients)
}

func TestConnectionStore_SubscriptionUpdates(t *testing.T) {
	store := setupTestConnectionStore()
	client := createMockClient("client-1")

	store.AddClient(client)

	retrieved, ok := store.GetClient("client-1")
	require.True(t, ok)

	retrieved.Subscriptions["task.created"] = true
	retrieved.Subscriptions["task.updated"] = true

	clients := store.GetClientsByTopic("task.created")
	assert.Equal(t, 1, len(clients))
	assert.Equal(t, "client-1", clients[0].ID)

	retrieved.Subscriptions["contributor.created"] = true
	clients = store.GetClientsByTopic("contributor.created")
	assert.Equal(t, 1, len(clients))

	delete(retrieved.Subscriptions, "task.created")
	clients = store.GetClientsByTopic("task.created")
	assert.Equal(t, 0, len(clients))
}

func TestConnectionStore_ClientActivityUpdate(t *testing.T) {
	store := setupTestConnectionStore()
	client := createMockClient("client-1")
	initialTime := client.LastActiveAt

	store.AddClient(client)

	time.Sleep(10 * time.Millisecond)

	retrieved, ok := store.GetClient("client-1")
	require.True(t, ok)

	retrieved.LastActiveAt = time.Now()

	updated, ok := store.GetClient("client-1")
	require.True(t, ok)
	assert.True(t, updated.LastActiveAt.After(initialTime))
}

package repository

import (
	"log/slog"
	"sync"

	"github.com/gocasters/rankr/realtimeapp/service/realtime"
)

type ConnectionStore struct {
	Clients map[string]*realtime.Client
	Mu      sync.RWMutex
	Logger  *slog.Logger
}

func NewConnectionStore(logger *slog.Logger) *ConnectionStore {
	return &ConnectionStore{
		Clients: make(map[string]*realtime.Client),
		Logger:  logger,
	}
}

func (cs *ConnectionStore) AddClient(client *realtime.Client) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	cs.Clients[client.ID] = client
	cs.Logger.Info("client added to store", "client_id", client.ID, "total_clients", len(cs.Clients))
}

func (cs *ConnectionStore) RemoveClient(clientID string) {
	cs.Mu.Lock()
	defer cs.Mu.Unlock()
	delete(cs.Clients, clientID)
	cs.Logger.Info("client removed from store", "client_id", clientID, "total_clients", len(cs.Clients))
}

func (cs *ConnectionStore) GetClient(clientID string) (*realtime.Client, bool) {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()
	client, ok := cs.Clients[clientID]
	return client, ok
}

func (cs *ConnectionStore) GetAllClients() []*realtime.Client {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()

	clients := make([]*realtime.Client, 0, len(cs.Clients))
	for _, client := range cs.Clients {
		clients = append(clients, client)
	}
	return clients
}

func (cs *ConnectionStore) GetClientsByTopic(topic string) []*realtime.Client {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()

	clients := make([]*realtime.Client, 0)
	for _, client := range cs.Clients {
		if client.Subscriptions[topic] {
			clients = append(clients, client)
		}
	}
	return clients
}

func (cs *ConnectionStore) ClientCount() int {
	cs.Mu.RLock()
	defer cs.Mu.RUnlock()
	return len(cs.Clients)
}

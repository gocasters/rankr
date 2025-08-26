package websocket

import (
	"github.com/google/uuid"
	"sync"
)

type HubInterface interface {
	RegisterClient(client *Client)
	UnRegisterClient(client *Client)
	BroadcastToClients(uuids []uuid.UUID, message []byte)
	GetClient(uuid uuid.UUID) (*Client, bool)
	IsClientRegistered(uuid uuid.UUID) bool
	Run()
	Close()
	GetRegisterChan() chan *Client
	GetUnregisterChan() chan *Client
}

type BroadcastMessage struct {
	UUID    []uuid.UUID
	Message []byte
}

type Hub struct {
	Clients             map[*Client]bool
	UUIDClient          map[uuid.UUID]*Client
	Register            chan *Client
	Unregister          chan *Client
	Broadcast           chan BroadcastMessage
	broadcastBufferSize uint
	mu                  sync.RWMutex
}

func NewHub(broadcastBufferSize uint) *Hub {
	return &Hub{
		Clients:             make(map[*Client]bool),
		UUIDClient:          make(map[uuid.UUID]*Client),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		Broadcast:           make(chan BroadcastMessage, broadcastBufferSize),
		broadcastBufferSize: broadcastBufferSize,
		mu:                  sync.RWMutex{},
	}
}

func (h *Hub) RegisterClient(client *Client) {

}

func (h *Hub) UnRegisterClient(client *Client) {

}

func (h *Hub) BroadcastToClients(uuids []uuid.UUID, message []byte) {

}

func (h *Hub) GetClient(uuid uuid.UUID) (*Client, bool) {
	return nil, false
}

func (h *Hub) IsClientRegistered(uuid uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.UUIDClient[uuid]
	return exists
}

func (h *Hub) Run() {}

// Close gracefully shuts down the Hub by closing all client connections
// and stopping the Hub's run loop.
func (h *Hub) Close() {

}

func (h *Hub) GetRegisterChan() chan *Client {
	if h != nil {
		return h.Register
	}
	return nil
}

func (h *Hub) GetUnregisterChan() chan *Client {
	if h != nil {
		return h.Unregister
	}
	return nil
}

package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/topicsname"
	"github.com/gocasters/rankr/realtimeapp/service/realtime"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockConnectionStore struct {
	clients map[string]*realtime.Client
	mu      sync.RWMutex
}

func NewMockConnectionStore() *MockConnectionStore {
	return &MockConnectionStore{
		clients: make(map[string]*realtime.Client),
	}
}

func (m *MockConnectionStore) AddClient(client *realtime.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client
}

func (m *MockConnectionStore) RemoveClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, clientID)
}

func (m *MockConnectionStore) GetClient(clientID string) (*realtime.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[clientID]
	return client, ok
}

func (m *MockConnectionStore) GetAllClients() []*realtime.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]*realtime.Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients
}

func (m *MockConnectionStore) GetClientsByTopic(topic string) []*realtime.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]*realtime.Client, 0)
	for _, client := range m.clients {
		client.SubsMu.RLock()
		isSubscribed := client.Subscriptions[topic]
		client.SubsMu.RUnlock()

		if isSubscribed {
			clients = append(clients, client)
		}
	}
	return clients
}

func setupTestHandler() (Handler, realtime.Service, *MockConnectionStore) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockStore := NewMockConnectionStore()
	service := realtime.NewService(mockStore, logger)
	handler := NewHandler(service, logger)
	return handler, service, mockStore
}

func TestNewHandler(t *testing.T) {
	handler, _, _ := setupTestHandler()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.RealtimeService)
	assert.NotNil(t, handler.Logger)
}

func TestHandler_GetStats(t *testing.T) {
	handler, service, _ := setupTestHandler()
	e := echo.New()

	for i := 0; i < 3; i++ {
		client := &realtime.Client{
			ID:            string(rune('0' + i)),
			Send:          make(chan []byte, 256),
			Subscriptions: make(map[string]bool),
			ConnectedAt:   time.Now(),
			LastActiveAt:  time.Now(),
		}
		service.RegisterClient(client)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.getStats(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(3), response["connected_clients"])
	assert.NotNil(t, response["timestamp"])
}

func TestHandler_GetStats_NoClients(t *testing.T) {
	handler, _, _ := setupTestHandler()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/v1/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.getStats(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["connected_clients"])
}

func TestHandler_HandleWebSocket(t *testing.T) {
	handler, _, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	time.Sleep(100 * time.Millisecond)

	count := handler.RealtimeService.GetConnectedClientCount()
	assert.Equal(t, 1, count)
}

func TestHandler_WebSocketSubscribe(t *testing.T) {
	handler, _, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	subscribeMsg := realtime.Message{
		Type: topicsname.MessageTypeSubscribe,
		Payload: map[string]interface{}{
			"topics": []interface{}{"task.created", "task.updated"},
		},
	}

	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	assert.Equal(t, topicsname.MessageTypeAck, response["type"])
	payload, ok := response["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, payload["success"].(bool))
}

func TestHandler_WebSocketUnsubscribe(t *testing.T) {
	handler, _, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	subscribeMsg := realtime.Message{
		Type: topicsname.MessageTypeSubscribe,
		Payload: map[string]interface{}{
			"topics": []interface{}{"task.created", "task.updated"},
		},
	}
	err = ws.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	var subscribeResp map[string]interface{}
	err = ws.ReadJSON(&subscribeResp)
	require.NoError(t, err)

	unsubscribeMsg := realtime.Message{
		Type: topicsname.MessageTypeUnsubscribe,
		Payload: map[string]interface{}{
			"topics": []interface{}{"task.created"},
		},
	}
	err = ws.WriteJSON(unsubscribeMsg)
	require.NoError(t, err)

	var unsubscribeResp map[string]interface{}
	err = ws.ReadJSON(&unsubscribeResp)
	require.NoError(t, err)

	assert.Equal(t, topicsname.MessageTypeAck, unsubscribeResp["type"])
	payload, ok := unsubscribeResp["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, payload["success"].(bool))
}

func TestHandler_WebSocketInvalidMessage(t *testing.T) {
	handler, _, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	err = ws.WriteMessage(websocket.TextMessage, []byte("{invalid json}"))
	require.NoError(t, err)

	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	assert.Equal(t, topicsname.MessageTypeError, response["type"])
	assert.Contains(t, response["message"], "invalid message format")
}

func TestHandler_WebSocketUnknownMessageType(t *testing.T) {
	handler, _, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	unknownMsg := realtime.Message{
		Type:    "unknown_type",
		Payload: map[string]interface{}{},
	}
	err = ws.WriteJSON(unknownMsg)
	require.NoError(t, err)

	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	require.NoError(t, err)

	assert.Equal(t, topicsname.MessageTypeError, response["type"])
	assert.Contains(t, response["message"], "unknown message type")
}

func TestHandler_WebSocketPingPong(t *testing.T) {
	handler, service, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	ws.SetPongHandler(func(appData string) error {
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	count := service.GetConnectedClientCount()
	assert.Equal(t, 1, count)
}

func TestHandler_WebSocketConnectionClose(t *testing.T) {
	handler, service, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, service.GetConnectedClientCount())

	ws.Close()

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, service.GetConnectedClientCount())
}

func TestHandler_MultipleWebSocketClients(t *testing.T) {
	handler, service, _ := setupTestHandler()
	e := echo.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		handler.handleWebSocket(c)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	numClients := 5
	clients := make([]*websocket.Conn, numClients)
	for i := 0; i < numClients; i++ {
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		clients[i] = ws
		defer ws.Close()
	}

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, numClients, service.GetConnectedClientCount())

	for _, ws := range clients {
		subscribeMsg := realtime.Message{
			Type: topicsname.MessageTypeSubscribe,
			Payload: map[string]interface{}{
				"topics": []interface{}{"task.created"},
			},
		}
		err := ws.WriteJSON(subscribeMsg)
		require.NoError(t, err)

		var resp map[string]interface{}
		ws.ReadJSON(&resp)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestHandler_SendError(t *testing.T) {
	handler, _, _ := setupTestHandler()
	client := &realtime.Client{
		ID:            "test-client",
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}

	handler.sendError(client, "test error message")

	select {
	case msg := <-client.Send:
		var errorMsg realtime.ErrorMessage
		err := json.Unmarshal(msg, &errorMsg)
		require.NoError(t, err)
		assert.Equal(t, topicsname.MessageTypeError, errorMsg.Type)
		assert.Equal(t, "test error message", errorMsg.Message)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No error message received")
	}
}

func TestHandler_SendResponse(t *testing.T) {
	handler, _, _ := setupTestHandler()
	client := &realtime.Client{
		ID:            "test-client",
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}

	payload := realtime.SubscribeResponse{
		Success: true,
		Topics:  []string{"task.created"},
	}

	handler.sendResponse(client, topicsname.MessageTypeAck, payload)

	select {
	case msg := <-client.Send:
		var response map[string]interface{}
		err := json.Unmarshal(msg, &response)
		require.NoError(t, err)
		assert.Equal(t, topicsname.MessageTypeAck, response["type"])
		assert.NotNil(t, response["payload"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No response message received")
	}
}

func TestHandler_ChannelFullHandling(t *testing.T) {
	handler, _, _ := setupTestHandler()
	client := &realtime.Client{
		ID:            "test-client",
		Send:          make(chan []byte, 1),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}

	client.Send <- []byte("message1")

	done := make(chan bool)
	go func() {
		handler.sendError(client, "error message")
		done <- true
	}()

	select {
	case <-done:

	case <-time.After(100 * time.Millisecond):
		t.Fatal("sendError blocked when channel was full")
	}
}

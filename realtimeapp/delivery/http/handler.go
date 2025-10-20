package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gocasters/rankr/pkg/topicsname"
	"github.com/gocasters/rankr/realtimeapp/service/realtime"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
}

type Handler struct {
	RealtimeService realtime.Service
	Logger          *slog.Logger
}

func NewHandler(realtimeService realtime.Service, logger *slog.Logger) Handler {
	return Handler{
		RealtimeService: realtimeService,
		Logger:          logger,
	}
}

func (h Handler) handleWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.Logger.Error("failed to upgrade connection", "error", err)
		return err
	}

	clientID := uuid.New().String()
	client := &realtime.Client{
		ID:            clientID,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		ConnectedAt:   time.Now(),
		LastActiveAt:  time.Now(),
	}

	h.RealtimeService.RegisterClient(client)

	go h.writePump(client)
	go h.readPump(client)

	return nil
}

func (h Handler) readPump(client *realtime.Client) {
	defer func() {
		h.RealtimeService.UnregisterClient(client.ID)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.Logger.Error("websocket error", "client_id", client.ID, "error", err)
			}
			break
		}

		client.LastActiveAt = time.Now()

		var msg realtime.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			h.Logger.Error("failed to unmarshal message", "client_id", client.ID, "error", err)
			h.sendError(client, "invalid message format")
			continue
		}

		h.handleClientMessage(client, msg)
	}
}

func (h Handler) writePump(client *realtime.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.Logger.Error("failed to write message", "client_id", client.ID, "error", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h Handler) handleClientMessage(client *realtime.Client, msg realtime.Message) {
	ctx := context.Background()

	switch msg.Type {
	case topicsname.MessageTypeSubscribe:
		var req realtime.SubscribeRequest
		topicsData, _ := json.Marshal(msg.Payload["topics"])
		json.Unmarshal(topicsData, &req.Topics)

		resp, err := h.RealtimeService.SubscribeTopics(ctx, client.ID, req)
		if err != nil {
			h.Logger.Error("failed to subscribe topics", "client_id", client.ID, "error", err)
			h.sendError(client, "failed to subscribe")
			return
		}

		h.sendResponse(client, topicsname.MessageTypeAck, resp)

	case topicsname.MessageTypeUnsubscribe:
		var req realtime.UnsubscribeRequest
		topicsData, _ := json.Marshal(msg.Payload["topics"])
		json.Unmarshal(topicsData, &req.Topics)

		resp, err := h.RealtimeService.UnsubscribeTopics(ctx, client.ID, req)
		if err != nil {
			h.Logger.Error("failed to unsubscribe topics", "client_id", client.ID, "error", err)
			h.sendError(client, "failed to unsubscribe")
			return
		}

		h.sendResponse(client, topicsname.MessageTypeAck, resp)

	default:
		h.sendError(client, "unknown message type")
	}
}

func (h Handler) sendError(client *realtime.Client, message string) {
	errMsg := realtime.ErrorMessage{
		Type:    topicsname.MessageTypeError,
		Message: message,
	}

	data, err := json.Marshal(errMsg)
	if err != nil {
		h.Logger.Error("failed to marshal error message", "error", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		h.Logger.Warn("client send channel full, dropping error message", "client_id", client.ID)
	}
}

func (h Handler) sendResponse(client *realtime.Client, msgType string, payload interface{}) {
	response := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.Logger.Error("failed to marshal response", "error", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		h.Logger.Warn("client send channel full, dropping response", "client_id", client.ID)
	}
}

func (h Handler) getStats(c echo.Context) error {
	stats := map[string]interface{}{
		"connected_clients": h.RealtimeService.GetConnectedClientCount(),
		"timestamp":         time.Now(),
	}

	return c.JSON(http.StatusOK, stats)
}

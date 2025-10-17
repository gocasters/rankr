package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gocasters/rankr/pkg/topicsname"
)

type ConnectionStore interface {
	AddClient(client *Client)
	RemoveClient(clientID string)
	GetClient(clientID string) (*Client, bool)
	GetAllClients() []*Client
	GetClientsByTopic(topic string) []*Client
}

type Service struct {
	ConnectionStore ConnectionStore
	TopicValidator  *TopicValidator
	Logger          *slog.Logger
}

func NewService(
	connectionStore ConnectionStore,
	logger *slog.Logger,
) Service {
	return Service{
		ConnectionStore: connectionStore,
		TopicValidator:  NewTopicValidator(),
		Logger:          logger,
	}
}

func (s Service) RegisterClient(client *Client) {
	s.ConnectionStore.AddClient(client)
	s.Logger.Info("client registered", "client_id", client.ID)
}

func (s Service) UnregisterClient(clientID string) {
	if client, ok := s.ConnectionStore.GetClient(clientID); ok {
		close(client.Send)
	}
	s.ConnectionStore.RemoveClient(clientID)
	s.Logger.Info("client unregistered", "client_id", clientID)
}

func (s Service) SubscribeTopics(ctx context.Context, clientID string, req SubscribeRequest) (SubscribeResponse, error) {
	client, ok := s.ConnectionStore.GetClient(clientID)
	if !ok {
		s.Logger.Error("client not found", "client_id", clientID)
		return SubscribeResponse{
			Success: false,
			Message: "client not found",
		}, nil
	}

	// TODO: Get actual client permissions from authentication context
	clientPerms := DefaultClientPermissions()

	allowedTopics, _, err := s.TopicValidator.ValidateTopics(req.Topics, clientPerms)

	if err != nil {
		return SubscribeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	client.SubsMu.Lock()
	for _, topic := range allowedTopics {
		client.Subscriptions[topic] = true
	}
	client.SubsMu.Unlock()

	s.Logger.Info("client subscribed to topics",
		"client_id", clientID,
		"allowed_topics", allowedTopics,
	)

	response := SubscribeResponse{
		Success: len(allowedTopics) > 0,
		Topics:  allowedTopics,
	}

	if len(allowedTopics) == 0 {
		response.Message = "no topics allowed"
	}

	return response, nil
}

func (s Service) UnsubscribeTopics(ctx context.Context, clientID string, req UnsubscribeRequest) (UnsubscribeResponse, error) {
	client, ok := s.ConnectionStore.GetClient(clientID)
	if !ok {
		s.Logger.Error("client not found", "client_id", clientID)
		return UnsubscribeResponse{
			Success: false,
			Message: "client not found",
		}, nil
	}

	client.SubsMu.Lock()
	defer client.SubsMu.Unlock()

	for _, topic := range req.Topics {
		delete(client.Subscriptions, topic)
	}

	s.Logger.Info("client unsubscribed from topics", "client_id", clientID, "topics", req.Topics)

	return UnsubscribeResponse{
		Success: true,
		Topics:  req.Topics,
	}, nil
}

func (s Service) BroadcastEvent(ctx context.Context, req BroadcastEventRequest) error {
	event := Event{
		Type:      topicsname.MessageTypeEvent,
		Topic:     req.Topic,
		Payload:   req.Payload,
		Timestamp: time.Now(),
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		s.Logger.Error("failed to marshal event", "error", err)
		return err
	}

	clients := s.ConnectionStore.GetClientsByTopic(req.Topic)
	s.Logger.Info("broadcasting event", "topicnames", req.Topic, "client_count", len(clients))

	for _, client := range clients {
		select {
		case client.Send <- eventData:
			// Event sent successfully
		default:
			// Client's send channel is full, skip
			s.Logger.Warn("client send channel full, skipping", "client_id", client.ID)
		}
	}

	return nil
}

func (s Service) GetConnectedClientCount() int {
	return len(s.ConnectionStore.GetAllClients())
}

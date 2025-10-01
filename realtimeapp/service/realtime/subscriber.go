package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Subscriber struct {
	Subscriber message.Subscriber
	Service    Service
	Logger     *slog.Logger
	Topics     []string
}

func NewSubscriber(
	subscriber message.Subscriber,
	service Service,
	topics []string,
	logger *slog.Logger,
) *Subscriber {
	return &Subscriber{
		Subscriber: subscriber,
		Service:    service,
		Topics:     topics,
		Logger:     logger,
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	for _, topic := range s.Topics {
		messages, err := s.Subscriber.Subscribe(ctx, topic)
		if err != nil {
			s.Logger.Error("failed to subscribe to topic", "topic", topic, "error", err)
			return err
		}

		go s.processMessages(ctx, topic, messages)
		s.Logger.Info("subscribed to topic", "topic", topic)
	}

	return nil
}

func (s *Subscriber) processMessages(ctx context.Context, topic string, messages <-chan *message.Message) {
	for {
		select {
		case <-ctx.Done():
			s.Logger.Info("stopping message processing", "topic", topic)
			return
		case msg, ok := <-messages:
			if !ok {
				s.Logger.Warn("message channel closed", "topic", topic)
				return
			}

			s.handleMessage(ctx, topic, msg)
			msg.Ack()
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, topic string, msg *message.Message) {
	s.Logger.Info("received message from NATS", "topic", topic, "message_id", msg.UUID)

	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		s.Logger.Error("failed to unmarshal message payload", "error", err)
		return
	}

	req := BroadcastEventRequest{
		Topic:   topic,
		Payload: payload,
	}

	if err := s.Service.BroadcastEvent(ctx, req); err != nil {
		s.Logger.Error("failed to broadcast event", "topic", topic, "error", err)
	}
}

func (s *Subscriber) Stop() error {
	s.Logger.Info("stopping NATS subscriber")
	return s.Subscriber.Close()
}

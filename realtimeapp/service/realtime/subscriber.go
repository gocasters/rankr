package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
)

// ServiceInterface defines the methods required by the Subscriber
type ServiceInterface interface {
	BroadcastEvent(ctx context.Context, req BroadcastEventRequest) error
}

type Subscriber struct {
	Subscriber message.Subscriber
	Service    ServiceInterface
	Logger     *slog.Logger
	Topics     []string
	wg         sync.WaitGroup
}

func NewSubscriber(
	subscriber message.Subscriber,
	service ServiceInterface,
	topics []string,
	logger *slog.Logger,
) *Subscriber {
	return &Subscriber{
		Subscriber: subscriber,
		Service:    service,
		Topics:     topics,
		Logger:     logger,
		wg:         sync.WaitGroup{},
	}
}

func (s *Subscriber) Start(ctx context.Context) error {
	for _, topic := range s.Topics {
		messages, err := s.Subscriber.Subscribe(ctx, topic)
		if err != nil {
			s.Logger.Error("failed to subscribe to topic", "topic", topic, "error", err)
			return err
		}

		s.wg.Add(1)

		go func(t string, msgs <-chan *message.Message) {
			defer s.wg.Done()
			s.processMessages(ctx, t, msgs)
		}(topic, messages)

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

			if err := s.handleMessage(ctx, topic, msg); err != nil {
				s.Logger.Error("failed to handle message", "topic", topic, "error", err)
				msg.Nack()
				continue
			} else {
				msg.Ack()

			}
		}
	}
}

func (s *Subscriber) handleMessage(ctx context.Context, topic string, msg *message.Message) error {
	s.Logger.Info("received message from NATS", "topic", topic, "message_id", msg.UUID)

	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		s.Logger.Error("failed to unmarshal message payload", "error", err)
		return err
	}

	req := BroadcastEventRequest{
		Topic:   topic,
		Payload: payload,
	}

	if err := s.Service.BroadcastEvent(ctx, req); err != nil {
		s.Logger.Error("failed to broadcast event", "topic", topic, "error", err)
		return err
	}
	return nil
}

func (s *Subscriber) Stop() error {
	s.Logger.Info("stopping NATS subscriber")
	err := s.Subscriber.Close()
	s.wg.Wait()
	return err
}

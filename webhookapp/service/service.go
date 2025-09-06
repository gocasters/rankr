package service

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"google.golang.org/protobuf/proto"
)

type EventRepository interface {
	Save(ctx context.Context, event *eventpb.Event) error
}

type Service struct {
	repo      EventRepository
	publisher message.Publisher
}

func New(repo EventRepository, publisher message.Publisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *Service) publishEvent(ev *eventpb.Event, evName eventpb.EventName, topic Topic, metadata map[string]string) error {
	payload, err := proto.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf event. eventname: %s. error: %w",
			evName, err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	fmt.Printf("event %s published to %s\n", evName, topic)

	sErr := s.repo.Save(context.Background(), ev)
	if sErr != nil {
		return fmt.Errorf("failed to save event. eventname: %s, error: %w",
			evName, err)
	}

	if err := s.publisher.Publish(string(topic), msg); err != nil {
		return fmt.Errorf("failed to publish event. topic: %s, eventname: %s, error: %w",
			topic, evName, err)
	}

	return nil
}

package publishevent

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service"
	"google.golang.org/protobuf/proto"
)

type EventRepository interface {
	Save(ctx context.Context, event *eventpb.Event) error
}

type RawEventRepository interface {
	Save(ctx context.Context, event *service.RawEvent) error
}

type Service struct {
	repo      EventRepository
	rawRepo   RawEventRepository
	publisher message.Publisher
}

// New creates a Service that persists events using repo and publishes them with publisher.
// The returned *Service coordinates saving events via the provided EventRepository and RawEventRepository and
// publishing Watermill messages using the provided message.Publisher.
func New(repo EventRepository, rawRepo RawEventRepository, publisher message.Publisher) *Service {
	return &Service{
		repo:      repo,
		rawRepo:   rawRepo,
		publisher: publisher,
	}
}

func (s *Service) publishEvent(ev *eventpb.Event, evName eventpb.EventName, topic service.Topic, metadata map[string]string) error {
	sErr := s.repo.Save(context.Background(), ev)
	if sErr != nil {
		return fmt.Errorf("failed to save event. eventname: %s, error: %w",
			evName, sErr)
	}

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

	if err := s.publisher.Publish(string(topic), msg); err != nil {
		return fmt.Errorf("failed to publish event. topic: %s, eventname: %s, error: %w",
			topic, evName, err)
	}

	return nil
}

func (s *Service) SaveRawEvent(event *service.RawEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if err := s.rawRepo.Save(context.Background(), event); err != nil {
		return fmt.Errorf("failed to save raw event. provider: %s, deliveryID: %s, error: %w",
			event.Provider, event.DeliveryID, err)
	}
	return nil
}

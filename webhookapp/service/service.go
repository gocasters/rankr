package service

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	Publisher message.Publisher
}

func New(publisher message.Publisher) *Service {
	return &Service{
		Publisher: publisher,
	}
}

func (s *Service) publishEvent(ev *eventpb.Event, evName eventpb.EventName, topic EventType, metadata map[string]string) error {
	payload, err := proto.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf event. eventname: %s. error: %w",
			evName, err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	if err := s.Publisher.Publish(string(topic), msg); err != nil {
		return fmt.Errorf("failed to publish event. topic: %s, eventname: %s, error: %w",
			topic, evName, err)
	}

	return nil
}

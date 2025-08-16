package event

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher struct {
	pub message.Publisher
}

func NewPublisher(pub message.Publisher) *Publisher {
	return &Publisher{pub: pub}
}

func (e *Publisher) Publish(ctx context.Context, topic string, event interface{}, metadata map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	if err := e.pub.Publish(topic, msg); err != nil {
		return err
	}

	return nil
}

func (e *Publisher) Close() error {
	return e.pub.Close()
}

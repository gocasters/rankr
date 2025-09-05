package broker

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher interface {
	Publish(topic string, msg *message.Message) error
	Close() error
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
	Close() error
}

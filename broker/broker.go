package broker

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher interface {
	Publish(topic string, msg *message.Message) error
}

type Subscriber interface {
	Subscribe(topic string) (<-chan *message.Message, error)
}

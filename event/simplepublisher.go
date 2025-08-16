package event

import (
	"fmt"
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
)

type SimplePublisher struct {
	closed bool
}

func (p *SimplePublisher) Publish(topic string, messages ...*message.Message) error {
	if p.closed {
		return fmt.Errorf("publisher is closed")
	}

	for _, msg := range messages {
		log.Println("Published message ", topic, msg.UUID, string(msg.Payload))
	}

	return nil
}

func (p *SimplePublisher) Close() error {
	p.closed = true
	log.Println("SimplePublisher closed")
	return nil
}

package event

import "context"

type Publisher interface {
	Publish(event Event) error
}

type Consumer interface {
	Consume(ctx context.Context, out chan<- Event) error
	// Consume(chan<- Event) error
}

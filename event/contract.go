package event

type Publisher interface {
	Publish(event Event) error
}

type Consumer interface {
	Consume(chan<- Event) error
}

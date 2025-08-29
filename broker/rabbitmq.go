package broker

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type RabbitBroker struct {
	publisher  message.Publisher
	subscriber *amqp.Subscriber
}

// amqpURL:  ( amqp://guest:guest@localhost:5672/)

func NewRabbitBroker(amqpURL, exchange, queue string, logger watermill.LoggerAdapter) *RabbitBroker {
	pub, err := amqp.NewPublisher(
		amqp.Config{
			Connection: amqp.ConnectionConfig{
				AmqpURI: amqpURL,
			},
			Exchange:  amqp.ExchangeConfig{GenerateName: amqp.GenerateQueueNameConstant(exchange)},
			Marshaler: amqp.DefaultMarshaler{},
		}, logger)
	if err != nil {
		panic(err)
	}

	sub, err := amqp.NewSubscriber(
		amqp.Config{
			Connection: amqp.ConnectionConfig{
				AmqpURI: amqpURL,
			},
			Queue:      amqp.QueueConfig{GenerateName: amqp.GenerateQueueNameConstant(queue)},
			Consume: amqp.ConsumeConfig{Consumer: "consumer1"},
			Marshaler:  amqp.DefaultMarshaler{},
		}, logger)
	if err != nil {
		panic(err)
	}

	return &RabbitBroker{
		publisher:  pub,
		subscriber: sub,
	}
}

func (r *RabbitBroker) Publish(topic string, msg *message.Message) error {
	return r.publisher.Publish(topic, msg)
}

func (r *RabbitBroker) Subscribe(topic string) (<-chan *message.Message, error) {
	return r.subscriber.Subscribe(context.Background(), topic)
}

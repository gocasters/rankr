package broker

import (
	"context"
	"errors"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type RabbitBroker struct {
	Publisher  message.Publisher
	Subscriber message.Subscriber
}

// amqpURL:  ( amqp://guest:guest@localhost:5672/)

func NewRabbitBroker(amqpURL, exchange, queue string, logger watermill.LoggerAdapter)  (*RabbitBroker, error) {
	pub, err := amqp.NewPublisher(
		amqp.Config{
			Connection: amqp.ConnectionConfig{
				AmqpURI: amqpURL,
			},
			Exchange:  amqp.ExchangeConfig{GenerateName: amqp.GenerateQueueNameConstant(exchange)},
			Marshaler: amqp.DefaultMarshaler{},
		}, logger)
	if err != nil {
		return nil ,err
	}

	sub, err := amqp.NewSubscriber(
		amqp.Config{
			Connection: amqp.ConnectionConfig{
				AmqpURI: amqpURL,
			},
			Queue:      amqp.QueueConfig{GenerateName: func(_ string) string { return queue }},
			Exchange:   amqp.ExchangeConfig{GenerateName: func(_ string) string { return exchange }},
			Consume: amqp.ConsumeConfig{Consumer: "consumer1"},
			Marshaler:  amqp.DefaultMarshaler{},
		}, logger)
	if err != nil {
		return nil,err
	}

	return &RabbitBroker{
		Publisher:  pub,
		Subscriber: sub,
	},nil
}


//pubish multi msg
func (r *RabbitBroker) Publish(topic string, msgs ...*message.Message) error {
	return r.Publisher.Publish(topic, msgs...)
}



func (r *RabbitBroker) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return r.Subscriber.Subscribe(ctx, topic)
}

func (r *RabbitBroker) Close() error {
      var err error
    if e := r.Publisher.Close(); e != nil {
        err = errors.Join(err, e)
    }
    if e := r.Subscriber.Close(); e != nil {
        err = errors.Join(err, e)
    }
   return err
}


// func (r *RabbitBroker) Publish(topic string, msg *message.Message) error {
// 	return r.publisher.Publish(topic, msg)
// }

// func (r *RabbitBroker) Subscribe(topic string) (<-chan *message.Message, error) {
// 	return r.subscriber.Subscribe(context.Background(), topic)
// }

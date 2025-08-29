package broker

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

type KafkaBroker struct {
	publisher  message.Publisher
	subscriber message.Subscriber
}

func NewKafkaBroker(brokers []string, group string, logger watermill.LoggerAdapter) *KafkaBroker {
	
	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:   brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}, logger)
	if err != nil {
		panic(err)
	}

	sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:       brokers,
		ConsumerGroup: group,
		Unmarshaler:   kafka.DefaultMarshaler{},
	}, logger)
	if err != nil {
		panic(err)
	}

	return &KafkaBroker{
		publisher:  pub,
		subscriber: sub,
	}
}

func (k *KafkaBroker) Publish(topic string, msg *message.Message) error {
	return k.publisher.Publish(topic, msg)
}

func (k *KafkaBroker) Subscribe(topic string) (<-chan *message.Message, error) {
	ch := make(chan *message.Message)

	go func() {
		subCh, err := k.subscriber.Subscribe(context.Background(), topic)
		if err != nil {
			panic(err)
		}
		for msg := range subCh {
			ch <- msg
		}
	}()

	return ch, nil
}


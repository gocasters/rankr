package broker

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

type KafkaBroker struct {
	Publisher  message.Publisher
	Subscriber message.Subscriber
}

func NewKafkaBroker(brokers []string, group string, logger watermill.LoggerAdapter) (*KafkaBroker,error) {
	
	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:   brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}, logger)
	if err != nil {
		return nil, err
	}


	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:       brokers,
		ConsumerGroup: group,
		Unmarshaler:   kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: config,
	}, logger)

	if err != nil {
		return nil, err
	}

	return &KafkaBroker{
		Publisher:  pub,
		Subscriber: sub,
	},nil
}

func (k *KafkaBroker) Publish(topic string, msg *message.Message) error {
	return k.Publisher.Publish(topic, msg)
}



func (k *KafkaBroker) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return k.Subscriber.Subscribe(ctx, topic)
}

func (k *KafkaBroker) Close() error {
    if err := k.Publisher.Close(); err != nil {
        return err
    }
    return k.Subscriber.Close()
}


 // func (k *KafkaBroker) Subscribe(topic string) (<-chan *message.Message, error) {
// 	ch := make(chan *message.Message)

// 	go func() {
// 		subCh, err := k.subscriber.Subscribe(context.Background(), topic)
// 		if err != nil {
// 			panic(err)
// 		}
// 		for msg := range subCh {
// 			ch <- msg
// 		}
// 	}()

// 	return ch, nil
// }
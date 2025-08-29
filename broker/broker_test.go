package broker

import (
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/stretchr/testify/require"
)

func TestKafkaBroker(t *testing.T) {
	logger := watermill.NewStdLogger(false, false)
	brokers := []string{"localhost:9092"}
	group := "test-group"

	kafkaBroker := NewKafkaBroker(brokers, group, logger)

	msg := message.NewMessage(watermill.NewUUID(), []byte("hello kafka"))
	err := kafkaBroker.Publish("test-topic", msg)
	require.NoError(t, err)

	ch, err := kafkaBroker.Subscribe("test-topic")
	require.NoError(t, err)

	select {
	case m := <-ch:
		require.Equal(t, "hello kafka", string(m.Payload))
	case <-time.After(5 * time.Second):
		t.Fatal("no message received from Kafka")
	}
}

func TestRabbitBroker(t *testing.T) {
	logger := watermill.NewStdLogger(false, false)
	amqpURL := "amqp://guest:guest@localhost:5672/"
	queue := "test-queue"
	exchange := "test-exchange"

	rabbitBroker := NewRabbitBroker(amqpURL, exchange, queue, logger)

	msg := message.NewMessage(watermill.NewUUID(), []byte("hello rabbit"))
	err := rabbitBroker.Publish(queue, msg)
	require.NoError(t, err)

	ch, err := rabbitBroker.Subscribe(queue)
	require.NoError(t, err)

	select {
	case m := <-ch:
		require.Equal(t, "hello rabbit", string(m.Payload))
	case <-time.After(5 * time.Second):
		t.Fatal("no message received from RabbitMQ")
	}
}

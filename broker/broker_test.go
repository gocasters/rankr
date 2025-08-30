package broker_test

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/broker"
	"github.com/stretchr/testify/assert"
)

type mockPublisher struct {
	published []*message.Message
	err       error
}

func (m *mockPublisher) Publish(topic string, msgs ...*message.Message) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, msgs...)
	return nil
}
func (m *mockPublisher) Close() error { return nil }

type mockSubscriber struct {
	msgs chan *message.Message
	err  error
}

func (m *mockSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.msgs, nil
}
func (m *mockSubscriber) Close() error { return nil }

func TestKafkaBroker_Publish(t *testing.T) {
	mockPub := &mockPublisher{}
	mockSub := &mockSubscriber{msgs: make(chan *message.Message, 1)}

	b := &broker.KafkaBroker{
		Publisher:  mockPub,
		Subscriber: mockSub,
	}

	msg := message.NewMessage("1", []byte("hello"))
	err := b.Publish("test-topic", msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mockPub.published))
	assert.Equal(t, "hello", string(mockPub.published[0].Payload))
}

func TestKafkaBroker_Subscribe(t *testing.T) {
	mockPub := &mockPublisher{}
	mockSub := &mockSubscriber{msgs: make(chan *message.Message, 1)}

	b := &broker.KafkaBroker{
		Publisher:  mockPub,
		Subscriber: mockSub,
	}

	expectedMsg := message.NewMessage("2", []byte("world"))
	mockSub.msgs <- expectedMsg

	ch, err := b.Subscribe(context.Background(), "test-topic")
	assert.NoError(t, err)
	msg := <-ch
	assert.Equal(t, "world", string(msg.Payload))
}

func TestRabbitBroker_PublishAndSubscribe(t *testing.T) {
	mockPub := &mockPublisher{}
	mockSub := &mockSubscriber{msgs: make(chan *message.Message, 1)}

	b := &broker.RabbitBroker{
		Publisher:  mockPub,
		Subscriber: mockSub,
	}

	msg := message.NewMessage("3", []byte("rabbit"))
	_ = b.Publish("rabbit-topic", msg)

	ch, _ := b.Subscribe(context.Background(), "rabbit-topic")
	mockSub.msgs <- msg

	got := <-ch
	assert.Equal(t, "rabbit", string(got.Payload))
}

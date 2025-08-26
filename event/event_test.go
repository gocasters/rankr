package event_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gocasters/rankr/event"
	"github.com/gocasters/rankr/pkg/logger"
)

type MockConsumer struct {
	Events []event.Event
	Err    error
}

func (m *MockConsumer) Consume(ctx context.Context, ch chan<- event.Event) error {
	if m.Err != nil {
		return m.Err
	}
	go func() {
		for _, e := range m.Events {
			select {
			case <-ctx.Done():
				return
			case ch <- e:
			}
		}
	}()
	return nil
}

func TestMain(m *testing.M) {
	cfg := logger.Config{
		Level:            "debug",
		FilePath:         "test.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	logger.Init(cfg)
	m.Run()
}

func TestEventConsumer_Start(t *testing.T) {
	ev1 := event.Event{Topic: "topic1", Payload: []byte("data1")}
	ev2 := event.Event{Topic: "topic2", Payload: []byte("data2")}

	var topic1Handled, topic2Handled bool
	var wg sync.WaitGroup
	wg.Add(2)

	router := event.Router{
		"topic1": func(e event.Event) error {
			topic1Handled = true
			wg.Done()
			return nil
		},
		"topic2": func(e event.Event) error {
			topic2Handled = true
			wg.Done()
			return errors.New("handler error")
		},
	}

	mock := &MockConsumer{Events: []event.Event{ev1, ev2}}

	consumer := event.EvnetConsumer{
		Consumers: []event.Consumer{mock},
		Router:    router,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)

	// wait for both handlers to complete
	wg.Wait()

	if !topic1Handled {
		t.Error("topic1 was not handled")
	}
	if !topic2Handled {
		t.Error("topic2 was not handled")
	}
}

func TestEventConsumer_ConsumerError(t *testing.T) {
	mock := &MockConsumer{Err: errors.New("consumer error")}

	consumer := event.EvnetConsumer{
		Consumers: []event.Consumer{mock},
		Router:    make(event.Router),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	consumer.Start(ctx)
	// just ensure it doesn't panic or block
	time.Sleep(50 * time.Millisecond)
}

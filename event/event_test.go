package event_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gocasters/rankr/event"
	"github.com/gocasters/rankr/pkg/logger"
)

type MockConsumer struct {
	Events []event.Event
	Err    error
}

func (m *MockConsumer) Consume(ch chan<- event.Event) error {
	if m.Err != nil {
		return m.Err
	}
	go func() {
		for _, e := range m.Events {
			ch <- e
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

func TestEvnetConsumer_Start(t *testing.T) {
	log, _ := logger.L()
	_ = log 

	ev1 := event.Event{Topic: "topic1", Payload: []byte("data1")}
	ev2 := event.Event{Topic: "topic2", Payload: []byte("data2")}

	handled := make(map[string]bool)

	router := event.Router{
		"topic1": func(e event.Event) error {
			handled["topic1"] = true
			return nil
		},
		"topic2": func(e event.Event) error {
			handled["topic2"] = true
			return errors.New("handler error")
		},
	}

	mock := &MockConsumer{Events: []event.Event{ev1, ev2}}

	consumer := event.EvnetConsumer{
		Consumers: []event.Consumer{mock},
		Router:    router,
	}

	done := make(chan bool)
	consumer.Start(done)

	time.Sleep(100 * time.Millisecond)

	if !handled["topic1"] {
		t.Error("topic1 was not handled")
	}
	if !handled["topic2"] {
		t.Error("topic2 was not handled")
	}

	close(done)
	time.Sleep(50 * time.Millisecond)
}

func TestEvnetConsumer_ConsumerError(t *testing.T) {
	mock := &MockConsumer{Err: errors.New("consumer error")}

	consumer := event.EvnetConsumer{
		Consumers: []event.Consumer{mock},
		Router:    make(event.Router),
	}

	done := make(chan bool)
	consumer.Start(done)

	time.Sleep(50 * time.Millisecond)
	close(done)
}

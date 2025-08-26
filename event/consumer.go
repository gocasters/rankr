package event

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gocasters/rankr/pkg/logger"
)

type handlerfunc func(event Event) error
type Router map[Topic]handlerfunc

type EventConsumer struct {
	Consumers []Consumer
	Router    Router
}

func (c EventConsumer) Start(ctx context.Context) {
	log, err := logger.L()
	if err != nil {
		log = slog.Default()
		log.Warn("logger not initialized; using default slog logger", "error", err)
	}

	eventstream := make(chan Event, 1024)

	var wg sync.WaitGroup
	for _, consumer := range c.Consumers {
		wg.Add(1)
		go func(cons Consumer) {
			defer wg.Done()
			if err := cons.Consume(ctx, eventstream); err != nil {
				log.Error("can't start consuming events", "error", err)
			}
		}(consumer)
	}
	// Close the stream only after cancellation AND all producers have returned. by coderabbit
	go func() {
		<-ctx.Done()
		log.Info("shutting down event consumer")
		wg.Wait()
		close(eventstream)
	}()

	// Dispatcher  by coderabbit
	go func() {
		for {
			e, ok := <-eventstream
			if !ok {
				log.Info("event stream closed; dispatcher exiting")
				return
			}
			if handler, ok := c.Router[e.Topic]; ok {
				if err := handler(e); err != nil {
					log.Error("can't handle event", "topic", e.Topic, "error", err)
				}
			} else {
				log.Warn("no handler for event topic", "topic", e.Topic)
			}
		}
	}()
}

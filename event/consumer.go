package event

import (
	"context"
	"log/slog"

	"github.com/gocasters/rankr/pkg/logger"
)

type handlerfunc func(event Event) error
type Router map[Topic]handlerfunc

type EvnetConsumer struct {
	Consumers []Consumer
	Router    Router
}

func (c EvnetConsumer) Start(ctx context.Context) {
	log, err := logger.L()
	if err != nil {
		log = slog.Default()
		log.Warn("logger not initialized; using default slog logger", "error", err)
	}

	
	
	eventstream := make(chan Event, 1024)
    for _, consumer := range c.Consumers {
       go func(cons Consumer) {
           if err := cons.Consume(ctx, eventstream); err != nil {
                log.Error("can't start consuming events", "error", err)
            }
        }(consumer)
    }
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("shutting down event consumer")
				close(eventstream)
                return
			case e := <-eventstream: 
				if handler, ok := c.Router[e.Topic]; ok {
					if err := handler(e); err != nil {
						log.Error("can't handle event", "topic", e.Topic, "error", err)
					}
				} else {
					log.Warn("no handler for event topic", "topic", e.Topic)
				}
			}
		}
	}()
}
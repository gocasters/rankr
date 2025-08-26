package event

import (
	"github.com/gocasters/rankr/pkg/logger"
)

type handlerfunc func(event Event) error
type Router map[Topic]handlerfunc

type EvnetConsumer struct {
	Consumers []Consumer
	Router    Router
}

func (c EvnetConsumer) Start(done <-chan bool) {
	log, err := logger.L()
	if err != nil {
		panic(err)
	}

	eventstream := make(chan Event, 1024)
	for _, consumer := range c.Consumers {
		if err := consumer.Consume(eventstream); err != nil {
			log.Error("can't start consuming events", "error", err)
		}
		
	}
	go func() {
		for {
			select {
			case <-done:
				log.Info("shutting down event consumer")
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
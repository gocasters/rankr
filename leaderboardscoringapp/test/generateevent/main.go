package main

import (
	"context"
	"flag"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/pkg/eventgenerator"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

func main() {
	natsURL := flag.String("nats", "nats://localhost:4222", "NATS URL")
	count := flag.Int("count", 10, "Number of events to generate")
	topic := flag.String("topic", "raw_events", "Topic name")
	flag.Parse()

	ctx := context.Background()
	logger := watermill.NewStdLogger(true, true)

	config := nats.Config{
		URL:          *natsURL,
		ClientID:     "event-generator",
		UseJetStream: true,
	}

	adapter, err := nats.New(ctx, config, logger)
	if err != nil {
		log.Fatalf("Failed to create NATS adapter: %v", err)
	}
	defer adapter.Close()

	generator := eventgenerator.NewEventGenerator(1, "test-repo")
	userIDs := []uint64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120}
	events := generator.GenerateRandomEvents(*count, userIDs)

	log.Printf("Publishing %d events to topic: %s", len(events), *topic)

	for i, event := range events {
		data, err := proto.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event %d: %v", i, err)
			continue
		}

		msg := message.NewMessage(watermill.NewULID(), data)
		if err := adapter.Publish(*topic, msg); err != nil {
			log.Printf("Failed to publish event %d: %v", i, err)
			continue
		}

		log.Printf("[%d/%d] Published: %s", i+1, len(events), event.EventName)
		time.Sleep(300 * time.Millisecond)
	}

	log.Println("All events published successfully!")
}

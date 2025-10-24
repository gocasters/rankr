// Package main provides a simple command-line tool for generating and publishing
// random raw events into NATS JetStream.
//
// This utility is designed to feed test or demo data into the leaderboard scoring
// service by simulating various repository events (issues, pull requests, commits, etc.).
// It uses the `pkg/eventgenerator` package to create random event payloads and publishes
// them to a specified NATS JetStream topic.
//
// Usage example:
//
//	go run ./cmd/eventgenerator/main.go --nats nats://localhost:4222 --count 50 --topic stream.raw.events
//
// Flags:
//
//	--nats     NATS server URL (default: nats://localhost:4222)
//	--count    Number of random events to generate (default: 10)
//	--topic    NATS topic name to publish events to (default: stream.raw.events)
//
// Each event is serialized using protobuf and sent as a Watermill message to JetStream.
package main

import (
	"context"
	"flag"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/pkg/eventgenerator"
	"github.com/gocasters/rankr/pkg/topicsname"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

func main() {
	natsURL := flag.String("nats", "nats://localhost:4222", "NATS URL")
	count := flag.Int("count", 10, "Number of events to generate")
	topic := flag.String("topic", topicsname.StreamNameRawEvents, "Topic name")
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

	generator := eventgenerator.NewEventGenerator(1, "test-repo1")
	userIDs := []uint64{
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
		111, 112, 113, 114, 115, 116, 117, 118, 119, 120,
		121, 122, 123, 124, 125, 126, 127, 128, 129, 130,
		131, 132, 133, 134, 135, 136, 137, 138, 139, 140,
	}
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

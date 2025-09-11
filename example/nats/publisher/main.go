package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
)

type ContributionEvent struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ProjectID   string    `json:"project_id"`
	EventType   string    `json:"event_type"`
	Timestamp   time.Time `json:"timestamp"`
	Points      int       `json:"points"`
	Description string    `json:"description"`
}

func main() {

	wmLogger := watermill.NewStdLogger(false, false)

	config := nats.Config{
		URL:            "nats://127.0.0.1:4222",
		ClientID:       "example-publisher",
		ConnectTimeout: 10 * time.Second,
		ReconnectWait:  2 * time.Second,
		MaxReconnects:  -1,
		PingInterval:   2 * time.Minute,
		MaxPingsOut:    2,
		AllowReconnect: true,
		UseJetStream:   true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	adapter, err := nats.New(ctx, config, wmLogger)
	if err != nil {
		log.Fatalf("Failed to create NATS adapter: %v", err)
	}
	defer func(adapter *nats.Adapter) {
		err := adapter.Close()
		if err != nil {
			log.Fatalf("Failed to close NATS adapter: %v", err)
		}
	}(adapter)

	publisher := adapter.Publisher()

	fmt.Println("Starting to publish events...")

	for i := 1; i <= 10; i++ {
		event := ContributionEvent{
			ID:          fmt.Sprintf("contrib-%d", i),
			UserID:      fmt.Sprintf("user-%d", i%3+1),
			ProjectID:   fmt.Sprintf("project-%d", i%2+1),
			EventType:   "commit",
			Timestamp:   time.Now(),
			Points:      10 + (i * 5),
			Description: fmt.Sprintf("Sample contribution event #%d", i),
		}

		eventData, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			continue
		}

		msg := message.NewMessage(watermill.NewUUID(), eventData)

		msg.Metadata.Set("event_type", event.EventType)
		msg.Metadata.Set("user_id", event.UserID)
		msg.Metadata.Set("project_id", event.ProjectID)

		if err := publisher.Publish("CONTRIBUTION_REGISTERED", msg); err != nil {
			log.Printf("Failed to publish message: %v", err)
			continue
		}

		fmt.Printf("Published event #%d: %s (User: %s, Project: %s, Points: %d)\n",
			i, event.ID, event.UserID, event.ProjectID, event.Points)

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Finished publishing events")
}

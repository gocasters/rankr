package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/gocasters/rankr/adapter/nats"
)

// ContributionEvent represents a sample event
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
		ClientID:       "example-subscriber",
		DurableName:    "example-consumer",
		QueueGroup:     "example-group",
		AckWaitTimeout: 30 * time.Second,
		MaxInflight:    1024,
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

	subscriber := adapter.Subscriber()

	messages, err := subscriber.Subscribe(ctx, "CONTRIBUTION_REGISTERED")
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	fmt.Println("Starting to consume messages... (Press Ctrl+C to stop)")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		messageCount := 0
		totalPoints := 0

		for msg := range messages {
			messageCount++

			var event ContributionEvent
			if err := json.Unmarshal(msg.Payload, &event); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Nack()
				continue
			}

			totalPoints += event.Points

			fmt.Printf("Received event #%d:\n", messageCount)
			fmt.Printf("  ID: %s\n", event.ID)
			fmt.Printf("  User: %s\n", event.UserID)
			fmt.Printf("  Project: %s\n", event.ProjectID)
			fmt.Printf("  Type: %s\n", event.EventType)
			fmt.Printf("  Points: %d\n", event.Points)
			fmt.Printf("  Description: %s\n", event.Description)
			fmt.Printf("  Timestamp: %s\n", event.Timestamp.Format(time.RFC3339))

			if eventType := msg.Metadata.Get("event_type"); eventType != "" {
				fmt.Printf("  Metadata - Event Type: %s\n", eventType)
			}
			if userID := msg.Metadata.Get("user_id"); userID != "" {
				fmt.Printf("  Metadata - User ID: %s\n", userID)
			}
			if projectID := msg.Metadata.Get("project_id"); projectID != "" {
				fmt.Printf("  Metadata - Project ID: %s\n", projectID)
			}

			fmt.Printf("  Running Total Points: %d\n", totalPoints)
			fmt.Println("  ---")

			time.Sleep(500 * time.Millisecond)

			msg.Ack()
		}
	}()

	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	if err := subscriber.Close(); err != nil {
		log.Printf("Error closing subscriber: %v", err)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
)

func main() {

	wmLogger := watermill.NewStdLogger(false, false)

	config := nats.Config{
		URL:            "nats://127.0.0.1:4222",
		ClientID:       "simple-example",
		UseJetStream:   true,
		ConnectTimeout: 10 * time.Second,
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

	fmt.Println("NATS adapter created successfully!")
	fmt.Printf("Connected to: %s\n", config.URL)
	fmt.Printf("Client ID: %s\n", config.ClientID)
	fmt.Printf("JetStream enabled: %t\n", config.UseJetStream)

	publisher := adapter.Publisher()

	fmt.Println("\nTesting publisher...")

	msg := message.NewMessage(watermill.NewUUID(), []byte("Hello, NATS!"))
	msg.Metadata.Set("timestamp", time.Now().Format(time.RFC3339))
	msg.Metadata.Set("sender", "simple-example")

	if err := publisher.Publish("TEST_SIMPLE", msg); err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	fmt.Println("Message published successfully!")
	fmt.Printf("Message ID: %s\n", msg.UUID)
	fmt.Printf("Payload: %s\n", string(msg.Payload))

	fmt.Println("\nTesting subscriber...")

	config.DurableName = "simple-consumer"
	config.QueueGroup = "simple-group"
	config.AckWaitTimeout = 30 * time.Second

	subAdapter, err := nats.New(ctx, config, wmLogger)
	if err != nil {
		log.Fatalf("Failed to create subscriber adapter: %v", err)
	}
	defer func(subAdapter *nats.Adapter) {
		err := subAdapter.Close()
		if err != nil {
			log.Fatalf("Failed to close NATS subscriber: %v", err)
		}
	}(subAdapter)

	subSubscriber := subAdapter.Subscriber()

	messages, err := subSubscriber.Subscribe(ctx, "TEST_SIMPLE")
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	fmt.Println("Waiting for messages... (will timeout in 5 seconds)")

	select {
	case receivedMsg := <-messages:
		fmt.Println("Message received!")
		fmt.Printf("Message ID: %s\n", receivedMsg.UUID)
		fmt.Printf("Payload: %s\n", string(receivedMsg.Payload))
		fmt.Printf("Timestamp: %s\n", receivedMsg.Metadata.Get("timestamp"))
		fmt.Printf("Sender: %s\n", receivedMsg.Metadata.Get("sender"))

		receivedMsg.Ack()
		fmt.Println("Message acknowledged!")

	case <-time.After(5 * time.Second):
		fmt.Println("No message received within timeout")
	}

	fmt.Println("\nSimple example completed successfully!")
}

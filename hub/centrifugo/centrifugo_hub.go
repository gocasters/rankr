package hub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/centrifugal/centrifuge-go"
)

type CentrifugoHub struct {
	client *centrifuge.Client
}

// NewCentrifugoHub creates and connects a new Centrifugo client
func NewCentrifugoHub(url, token string) (*CentrifugoHub, error) {
	cfg := centrifuge.Config{}
	cfg.Token = token

	client := centrifuge.NewJsonClient(url, cfg)

	// Setup some basic logging
	client.OnConnecting(func(e centrifuge.ConnectingEvent) {
		fmt.Println("Connecting:", e.Code, e.Reason)
	})
	client.OnConnected(func(e centrifuge.ConnectedEvent) {
		fmt.Println("Connected to Centrifugo")
	})
	client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		fmt.Println("Disconnected:", e.Code, e.Reason)
	})

	// Connect
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("centrifugo connect: %w", err)
	}

	return &CentrifugoHub{client: client}, nil
}

// Publish sends JSON data to a channel
func (h *CentrifugoHub) Publish(ctx context.Context, channel string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = h.client.Publish(ctx, channel, payload)
	return err
}

// Subscribe listens to a channel and calls handler when messages arrive
func (h *CentrifugoHub) Subscribe(ctx context.Context, channel string, handler func([]byte)) error {
	sub, err := h.client.NewSubscription(channel)
	if err != nil {
		return err
	}

	sub.OnPublication(func(e centrifuge.PublicationEvent) {
		handler(e.Data)
	})

	if err := sub.Subscribe(); err != nil {
		return err
	}

	return nil
}

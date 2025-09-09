package hub

import (
	"context"
	"encoding/json"
	"log"

	"github.com/centrifugal/centrifuge-go"
)

type Hub interface {
	Publish(ctx context.Context, channel string, data any) error
}

type CentrifugoHub struct {
	client *centrifuge.Client
}

func NewCentrifugoHub(url, token string) *CentrifugoHub {
	config := centrifuge.Config{
		Token: token,
	}

	client := centrifuge.NewJsonClient(url, config)

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect to Centrifugo: %v", err)
	}

	return &CentrifugoHub{client: client}
}

func (h *CentrifugoHub) Publish(ctx context.Context, channel string, data any) error {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	h.Publish(ctx, channel, jsonData)

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
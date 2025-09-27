package serializedevent

import (
	"fmt"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/proto"
	"time"
)

// WebhookEventRow represents a complete row from the webhook_events table
type WebhookEventRow struct {
	ID         int64
	Provider   int32
	Owner      string //TODO add these parameters to the repository methods and eventspb
	Repo       string
	HookID     int64
	DeliveryID string
	EventType  int32
	Payload    []byte
	ReceivedAt time.Time
}

// ToEvent converts a WebhookEventRow to an eventpb.Event
func (row *WebhookEventRow) ToEvent() (*eventpb.Event, error) {
	var event eventpb.Event
	if err := proto.Unmarshal(row.Payload, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}

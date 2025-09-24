package service

import (
	"encoding/json"
	"time"
)

type RawEvent struct {
	Provider   int             `json:"provider"`
	Owner      string          `json:"owner"`
	Repo       string          `json:"repo"`
	HookID     int64           `json:"hook_id"`
	DeliveryID string          `json:"delivery_id"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
}

type DeliveryEvent struct {
	Id             int        `json:"id"`
	Guid           string     `json:"guid"`
	DeliveredAt    time.Time  `json:"delivered_at"`
	Redelivery     bool       `json:"redelivery"`
	Duration       float64    `json:"duration"`
	Status         string     `json:"status"`
	StatusCode     int        `json:"status_code"`
	Event          string     `json:"event"`
	Action         *string    `json:"action"`
	InstallationId *int       `json:"installation_id"`
	RepositoryId   *int       `json:"repository_id"`
	ThrottledAt    *time.Time `json:"throttled_at"`
}

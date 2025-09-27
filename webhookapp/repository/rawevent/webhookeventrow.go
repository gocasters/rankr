package rawevent

import (
	"encoding/json"
	"fmt"
	"time"
)

// WebhookEventRow represents a complete row from the raw_webhook_events table
type WebhookEventRow struct {
	ID          int64           `json:"id"`
	Provider    int32           `json:"provider"`
	Owner       string          `json:"owner"`
	Repo        string          `json:"repo"`
	HookID      int64           `json:"hook_id"`
	DeliveryID  string          `json:"delivery_id"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	ReceivedAt  time.Time       `json:"received_at"`
}

// ToMap converts the JSON payload to a map[string]interface{}
func (row *WebhookEventRow) ToMap() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(row.PayloadJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON payload: %w", err)
	}
	return result, nil
}

// GetPayloadString returns the JSON payload as a formatted string
func (row *WebhookEventRow) GetPayloadString() string {
	return string(row.PayloadJSON)
}

// HasField checks if a specific field exists in the JSON payload
func (row *WebhookEventRow) HasField(fieldPath string) (bool, error) {
	var temp interface{}
	if err := json.Unmarshal(row.PayloadJSON, &temp); err != nil {
		return false, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Simple field existence check - you can extend this for nested paths
	data, ok := temp.(map[string]interface{})
	if !ok {
		return false, nil
	}

	_, exists := data[fieldPath]
	return exists, nil
}

// GetField extracts a specific field from the JSON payload
func (row *WebhookEventRow) GetField(fieldPath string) (interface{}, error) {
	data, err := row.ToMap()
	if err != nil {
		return nil, err
	}

	value, exists := data[fieldPath]
	if !exists {
		return nil, fmt.Errorf("field %s not found", fieldPath)
	}

	return value, nil
}

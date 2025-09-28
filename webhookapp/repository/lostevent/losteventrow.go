package lostevent

// LostEventRow represents a complete row from the raw_webhook_events table
type LostEventRow struct {
	ID         int64
	Provider   int32
	Owner      string
	Repo       string
	HookID     int64
	DeliveryID string
}

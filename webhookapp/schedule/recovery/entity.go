package recovery

// WebhookDelivery represents a webhook delivery from GitHub
type WebhookDelivery struct {
	ID          int     `json:"id"`
	GUID        string  `json:"guid"`
	DeliveredAt string  `json:"delivered_at"`
	Redelivery  bool    `json:"redelivery"`
	Duration    float64 `json:"duration"`
	Status      string  `json:"status"`
	StatusCode  int     `json:"status_code"`
	Event       string  `json:"event"`
	Action      string  `json:"action"`
}
type MissingDelivery struct {
	WebhookDelivery
	HookID              int
	RecoveryAttempts    int
	LastRecoveryAttempt string
	Owner               string
	Repo                string
}

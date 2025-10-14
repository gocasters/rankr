package realtime

type SubscribeRequest struct {
	Topics []string `json:"topics"`
}

type SubscribeResponse struct {
	Success      bool              `json:"success"`
	Topics       []string          `json:"topics"`
	Message      string            `json:"message,omitempty"`
	DeniedTopics map[string]string `json:"denied_topics,omitempty"`
}

type UnsubscribeRequest struct {
	Topics []string `json:"topics"`
}

type UnsubscribeResponse struct {
	Success bool     `json:"success"`
	Topics  []string `json:"topics"`
	Message string   `json:"message,omitempty"`
}

type BroadcastEventRequest struct {
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}

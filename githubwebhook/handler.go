package githubwebhook

import (
	"encoding/json"
	"github.com/gocasters/rankr/event"
	"io"
	"net/http"
)

type GithubWebhookHTTPHandler struct {
	publisher *event.Publisher
}

func NewGithubWebhookHandler(pub *event.Publisher) *GithubWebhookHTTPHandler {
	return &GithubWebhookHTTPHandler{publisher: pub}
}

func (h *GithubWebhookHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hookID := r.Header.Get("X-GitHub-Hook-ID")
	eventName := r.Header.Get("X-GitHub-Event")
	deliveryUID := r.Header.Get("X-GitHub-Delivery")

	if hookID == "" || eventName == "" || deliveryUID == "" {
		http.Error(w, "Missing required GitHub headers", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	acEvent := ActivityEvent{
		HookID:   hookID,
		Event:    eventName,
		Delivery: deliveryUID,
		Body:     json.RawMessage(body), // keep raw JSON
	}

	if err := h.publisher.Publish(r.Context(), TopicGithubUserActivity, acEvent, nil); err != nil {
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, wErr := w.Write([]byte(`{"status":"ok"}`))
	if wErr != nil {
		return
	}
}

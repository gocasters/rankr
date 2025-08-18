package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/githubwebhook"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

func (s Server) PublishGithubActivity(c echo.Context) error {
	hookID := c.Request().Header.Get("X-GitHub-Hook-ID")
	eventName := c.Request().Header.Get("X-GitHub-Event")
	deliveryUID := c.Request().Header.Get("X-GitHub-Delivery")

	if hookID == "" || eventName == "" || deliveryUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required GitHub headers",
		})
	}

	// TODO: add limit to read body to avoid abuse. -> "const maxPayload = 1 << 20";
	// TODO: body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxPayload))
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	webhookAction, waErr := extractWebhookAction(body)
	if waErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse JSON",
		})
	}

	// Map event types to their handler functions
	handlers := map[githubwebhook.EventType]func(ctx context.Context, action string, body []byte, uid string) error{
		githubwebhook.EventTypeIssues:            s.Service.HandleIssuesEvent,
		githubwebhook.EventTypePullRequest:       s.Service.HandlePullRequestEvent,
		githubwebhook.EventTypePullRequestReview: s.Service.HandlePullRequestReviewEvent,
	}

	if handler, ok := handlers[githubwebhook.EventType(eventName)]; ok {
		if eErr := handler(c.Request().Context(), webhookAction, body, deliveryUID); eErr != nil {
			if s.Handler.Logger != nil {
				s.Handler.Logger.Error("Failed to handle issues event",
					"err", err, "delivery", deliveryUID, "action", webhookAction)
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("failed to handle event. event type: %s", eventName),
			})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("event type %q not handled", eventName),
	})
}

func extractWebhookAction(body []byte) (string, error) {
	var actionData struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(body, &actionData); err != nil {
		return "", err
	}
	if actionData.Action == "" {
		return "", fmt.Errorf("missing action field")
	}
	return actionData.Action, nil
}

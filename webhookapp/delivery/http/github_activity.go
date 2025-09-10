package http

import (
	"encoding/json"
	"fmt"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

func (s *Server) PublishGithubActivity(c echo.Context) error {
	hookID := c.Request().Header.Get("X-GitHub-Hook-ID")
	eventName := c.Request().Header.Get("X-GitHub-Event")
	deliveryUID := c.Request().Header.Get("X-GitHub-Delivery")

	if err := validateGitHubHeaders(hookID, eventName, deliveryUID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
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

	var eventError error

	switch service.EventType(eventName) {
	case service.EventTypeIssues:
		eventError = s.Service.HandleIssuesEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case service.EventTypeIssueComment:
		eventError = s.Service.HandleIssueCommentEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case service.EventTypePullRequest:
		eventError = s.Service.HandlePullRequestEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case service.EventTypePullRequestReview:
		eventError = s.Service.HandlePullRequestReviewEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case service.EventTypePush:
		eventError = s.Service.HandlePushEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, body, deliveryUID)
	default:
		return c.JSON(http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Event types '%s' not handled", eventName),
		})
	}

	if eventError != nil {
		if s.Handler.Logger != nil {
			s.Handler.Logger.Error("Failed to handle event",
				"err", eventError, "event", eventName,
				"delivery", deliveryUID, "action", webhookAction)
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Failed to handle event. Event Type: %s", eventName),
		})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
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

func validateGitHubHeaders(hookID, eventName, deliveryUID string) error {
	if hookID == "" {
		return fmt.Errorf("missing X-GitHub-Hook-ID header")
	}
	if eventName == "" {
		return fmt.Errorf("missing X-GitHub-Event header")
	}
	if deliveryUID == "" {
		return fmt.Errorf("missing X-GitHub-Delivery header")
	}
	return nil
}

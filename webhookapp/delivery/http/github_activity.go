package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service/delivery"
	"github.com/labstack/echo/v4"
)

func (s *Server) PublishGithubActivity(c echo.Context) error {
	hookID := c.Request().Header.Get("X-GitHub-Hook-ID")
	eventName := c.Request().Header.Get("X-GitHub-Event")
	deliveryUID := c.Request().Header.Get("X-GitHub-Delivery")
	hubSig256 := c.Request().Header.Get("X-Hub-Signature-256")
	body, err := s.Service.VerifyEvent(c.Request().Body, hookID, eventName, deliveryUID, hubSig256)
	if err != nil {
		var httpStatus int
		if errors.Is(err, delivery.ErrWebhookHmacSigFailure) {
			httpStatus = http.StatusUnauthorized
		} else {
			httpStatus = http.StatusBadRequest
		}
		return c.JSON(httpStatus, map[string]string{
			"error": err.Error(),
		})
	}

	webhookAction, waErr := extractWebhookAction(body)
	if waErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse JSON",
		})
	}

	var eventError error

	switch delivery.EventType(eventName) {
	case delivery.EventTypeIssues:
		eventError = s.Service.HandleIssuesEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case delivery.EventTypeIssueComment:
		eventError = s.Service.HandleIssueCommentEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case delivery.EventTypePullRequest:
		eventError = s.Service.HandlePullRequestEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case delivery.EventTypePullRequestReview:
		eventError = s.Service.HandlePullRequestReviewEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, webhookAction, body, deliveryUID)
	case delivery.EventTypePush:
		eventError = s.Service.HandlePushEvent(eventpb.EventProvider_EVENT_PROVIDER_GITHUB, body, deliveryUID)
	default:
		return c.JSON(http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Event types '%s' not handled", eventName),
		})
	}

	if eventError != nil {
		logger.L().Error("Failed to handle event",
			"err", eventError, "event", eventName,
			"delivery", deliveryUID, "action", webhookAction)
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

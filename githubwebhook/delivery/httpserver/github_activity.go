package httpserver

import (
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

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	switch githubwebhook.EventType(eventName) {
	case githubwebhook.EventTypeIssues:
		webhookAction, waErr := extractWebhookAction(body)
		if waErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to parse JSON",
			})
		}

		handleEvent := s.Service.HandleIssuesEvent(c.Request().Context(), webhookAction, body, deliveryUID)
		if handleEvent != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to handle event",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})

	case githubwebhook.EventTypePullRequest:
		webhookAction, waErr := extractWebhookAction(body)
		if waErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to parse JSON",
			})
		}

		handleEvent := s.Service.HandlePullRequestEvent(c.Request().Context(), webhookAction, body, deliveryUID)
		if handleEvent != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to handle event",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})

	case githubwebhook.EventTypePullRequestReview:
		webhookAction, waErr := extractWebhookAction(body)
		if waErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to parse JSON",
			})
		}

		handleEvent := s.Service.HandlePullRequestReviewEvent(c.Request().Context(), webhookAction, body, deliveryUID)
		if handleEvent != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Failed to handle event",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})

	default:
		return c.JSON(http.StatusOK, map[string]string{
			"message": fmt.Sprintf("Event type '%s' not handled", eventName),
		})
	}
}

func extractWebhookAction(body []byte) (string, error) {
	var actionData struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(body, &actionData); err != nil {
		return "", err
	}
	return actionData.Action, nil
}

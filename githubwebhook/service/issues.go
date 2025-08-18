package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/githubwebhook"
)

func (s Service) HandleIssuesEvent(c context.Context, action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req githubwebhook.IssueOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.processIssueOpened(c, req, deliveryUID)

	case "closed":
		var req githubwebhook.IssueClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.processIssueClosed(c, req, deliveryUID)

	default:
		return fmt.Errorf("issue action '%s' not handled", action)
	}
}

func (s Service) processIssueOpened(c context.Context, req githubwebhook.IssueOpenedRequest, deliveryUID string) error {
	return s.publishActivityEvent(c, githubwebhook.EventTypeIssues,
		githubwebhook.PayloadTypeIssueOpened, req, deliveryUID)
}

func (s Service) processIssueClosed(c context.Context, req githubwebhook.IssueClosedRequest, deliveryUID string) error {
	return s.publishActivityEvent(c, githubwebhook.EventTypeIssues,
		githubwebhook.PayloadTypeIssueClosed, req, deliveryUID)
}

func (s Service) publishActivityEvent(ctx context.Context, eventType githubwebhook.EventType,
	payloadType githubwebhook.PaLoadType, payload interface{}, deliveryUID string) error {

	ev := githubwebhook.ActivityEvent{
		Event:       eventType,
		Delivery:    deliveryUID,
		PayloadType: payloadType,
		Payload:     payload,
	}

	metadata := make(map[string]string)
	if err := s.Publisher.Publish(ctx, githubwebhook.TopicGithubUserActivity, ev, metadata); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

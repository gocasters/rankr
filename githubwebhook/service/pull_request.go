package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/githubwebhook"
)

func (s Service) HandlePullRequestEvent(c context.Context, action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req githubwebhook.PullRequestOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		return s.ProcessPullRequestOpened(c, req, deliveryUID)

	case "closed":
		var req githubwebhook.PullRequestClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.ProcessPullRequestClosed(c, req, deliveryUID)

	default:
		return fmt.Errorf("pull request action '%s' not handled", action)
	}
}

func (s Service) ProcessPullRequestOpened(c context.Context, req githubwebhook.PullRequestOpenedRequest, deliveryUID string) error {
	ev := githubwebhook.ActivityEvent{
		Event:       githubwebhook.EventTypePullRequest,
		Delivery:    deliveryUID,
		PayloadType: githubwebhook.PayloadTypePullRequestOpened,
		Payload:     req,
	}

	metadata := make(map[string]string)

	if err := s.Publisher.Publish(c, githubwebhook.TopicGithubUserActivity, ev, metadata); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (s Service) ProcessPullRequestClosed(c context.Context, req githubwebhook.PullRequestClosedRequest, deliveryUID string) error {
	ev := githubwebhook.ActivityEvent{
		Event:       githubwebhook.EventTypePullRequest,
		Delivery:    deliveryUID,
		PayloadType: githubwebhook.PayloadTypePullRequestClosed,
		Payload:     req,
	}

	metadata := make(map[string]string)

	if err := s.Publisher.Publish(c, githubwebhook.TopicGithubUserActivity, ev, metadata); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/githubwebhook"
)

func (s Service) HandlePullRequestReviewEvent(c context.Context, action string, body []byte, deliveryUID string) error {
	switch action {
	case "submitted":
		var reviewData githubwebhook.PullRequestReviewSubmittedRequest
		if err := json.Unmarshal(body, &reviewData); err != nil {
			return err
		}
		return s.ProcessPullRequestReviewSubmitted(c, reviewData, deliveryUID)

	default:
		return fmt.Errorf("pull request review action '%s' not handled", action)
	}
}

func (s Service) ProcessPullRequestReviewSubmitted(c context.Context, req githubwebhook.PullRequestReviewSubmittedRequest, deliveryUID string) error {
	ev := githubwebhook.ActivityEvent{
		Event:       githubwebhook.EventTypePullRequestReview,
		Delivery:    deliveryUID,
		PayloadType: githubwebhook.PayloadTypePullRequestReviewSubmitted,
		Payload:     req,
	}

	metadata := make(map[string]string)

	if err := s.Publisher.Publish(c, githubwebhook.TopicGithubUserActivity, ev, metadata); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

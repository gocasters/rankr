package service

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (s *Service) HandlePullRequestReviewEvent(c context.Context, action string, body []byte, deliveryUID string) error {
	switch action {
	case "submitted":
		var reviewData PullRequestReviewSubmittedRequest
		if err := json.Unmarshal(body, &reviewData); err != nil {
			return err
		}
		return s.ProcessPullRequestReviewSubmitted(c, reviewData, deliveryUID)

	default:
		return fmt.Errorf("pull request review action '%s' not handled", action)
	}
}

func (s *Service) ProcessPullRequestReviewSubmitted(c context.Context, req PullRequestReviewSubmittedRequest, deliveryUID string) error {
	ev := ActivityEvent{
		Event:       EventTypePullRequestReview,
		Delivery:    deliveryUID,
		PayloadType: PayloadTypePullRequestReviewSubmitted,
		Payload:     req,
	}

	payload, pErr := json.Marshal(ev)
	if pErr != nil {
		return pErr
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	metadata := map[string]string{"delivery": deliveryUID}

	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	if err := s.Publisher.Publish(TopicGithubUserActivity, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

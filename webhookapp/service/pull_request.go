package service

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (s *Service) HandlePullRequestEvent(c context.Context, action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req PullRequestOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		return s.ProcessPullRequestOpened(c, req, deliveryUID)

	case "closed":
		var req PullRequestClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.ProcessPullRequestClosed(c, req, deliveryUID)

	default:
		return fmt.Errorf("pull request action '%s' not handled", action)
	}
}

func (s *Service) ProcessPullRequestOpened(c context.Context, req PullRequestOpenedRequest, deliveryUID string) error {
	ev := ActivityEvent{
		Event:       EventTypePullRequest,
		Delivery:    deliveryUID,
		PayloadType: PayloadTypePullRequestOpened,
		Payload:     req,
	}

	payload, pErr := json.Marshal(ev)
	if pErr != nil {
		return pErr
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	metadata := map[string]string{
		"delivery": deliveryUID,
		"action":   "opened",
	}

	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	if err := s.Publisher.Publish(TopicGithubUserActivity, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (s *Service) ProcessPullRequestClosed(c context.Context, req PullRequestClosedRequest, deliveryUID string) error {
	ev := ActivityEvent{
		Event:       EventTypePullRequest,
		Delivery:    deliveryUID,
		PayloadType: PayloadTypePullRequestClosed,
		Payload:     req,
	}

	payload, pErr := json.Marshal(ev)
	if pErr != nil {
		return pErr
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	metadata := map[string]string{
		"delivery": deliveryUID,
		"action":   "closed",
	}

	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	if err := s.Publisher.Publish(TopicGithubUserActivity, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

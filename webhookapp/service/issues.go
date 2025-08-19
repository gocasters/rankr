package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (s *Service) HandleIssuesEvent(c context.Context, action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req IssueOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.processIssueOpened(c, req, deliveryUID)

	case "closed":
		var req IssueClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.processIssueClosed(c, req, deliveryUID)

	default:
		return fmt.Errorf("issue action '%s' not handled", action)
	}
}

func (s *Service) processIssueOpened(c context.Context, req IssueOpenedRequest, deliveryUID string) error {
	return s.publishActivityEvent(c, EventTypeIssues,
		PayloadTypeIssueOpened, req, deliveryUID)
}

func (s *Service) processIssueClosed(c context.Context, req IssueClosedRequest, deliveryUID string) error {
	return s.publishActivityEvent(c, EventTypeIssues,
		PayloadTypeIssueClosed, req, deliveryUID)
}

func (s *Service) publishActivityEvent(ctx context.Context, eventType EventType,
	payloadType PayLoadType, payload interface{}, deliveryUID string) error {

	ev := ActivityEvent{
		Event:       eventType,
		Delivery:    deliveryUID,
		PayloadType: payloadType,
		Payload:     payload,
	}

	p, pErr := json.Marshal(ev)
	if pErr != nil {
		return pErr
	}

	msg := message.NewMessage(watermill.NewUUID(), p)

	metadata := map[string]string{"delivery": deliveryUID}

	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	if err := s.Publisher.Publish(TopicGithubUserActivity, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

package service

import (
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (s *Service) HandleIssuesEvent(action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req IssueOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.processIssueOpened(req, deliveryUID)

	case "closed":
		var req IssueClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.processIssueClosed(req, deliveryUID)

	default:
		return fmt.Errorf("issue action '%s' not handled", action)
	}
}

func (s *Service) processIssueOpened(req IssueOpenedRequest, deliveryUID string) error {
	return s.publishActivityEvent(EventTypeIssues,
		PayloadTypeIssueOpened, req, deliveryUID)
}

func (s *Service) processIssueClosed(req IssueClosedRequest, deliveryUID string) error {
	return s.publishActivityEvent(EventTypeIssues,
		PayloadTypeIssueClosed, req, deliveryUID)
}

func (s *Service) publishActivityEvent(eventType EventType,
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

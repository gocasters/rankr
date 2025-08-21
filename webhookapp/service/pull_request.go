package service

import (
	"encoding/json"
	"fmt"

	"github.com/gocasters/rankr/protobuf/golang/eventpb"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandlePullRequestEvent(action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req PullRequestOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		return s.ProcessPullRequestOpened(req, deliveryUID)

	case "closed":
		var req PullRequestClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.ProcessPullRequestClosed(req, deliveryUID)

	default:
		return fmt.Errorf("pull request action '%s' not handled", action)
	}
}

func (s *Service) ProcessPullRequestOpened(req PullRequestOpenedRequest, deliveryUID string) error {
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

func (s *Service) ProcessPullRequestClosed(req PullRequestClosedRequest, deliveryUID string) error {
	ev := &eventpb.Event{
		Id:        deliveryUID,
		EventName: eventpb.EventName_PULL_REQUEST_CLOSED,
		Time:      timestamppb.Now(),
		Payload: &eventpb.Event_PrClosedPayload{ // wrapper
			PrClosedPayload: &eventpb.PullRequestClosedPayload{
				UserId:    req.Sender.Login,
				ProjectId: req.Repository.FullName,
				Merged:    *req.PullRequest.Merged,
				Additions: int32(req.PullRequest.Additions),
				Deletions: int32(req.PullRequest.Deletions),
			},
		},
	}

	payload, err := proto.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
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

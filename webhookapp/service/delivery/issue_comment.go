package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandleIssueCommentEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) error {
	ev, err := buildIssueCommentEvent(provider, action, body, deliveryUID)
	if err != nil {
		return err
	}
	return s.saveEvent(context.Background(), ev)
}

func buildIssueCommentEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) (*eventpb.Event, error) {
	switch action {
	case "created":
		var req IssueCommentCreatedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, err
		}
		return newIssueCommentEvent(req, provider, deliveryUID), nil
	default:
		return nil, fmt.Errorf("issue_comment action '%s' not handled", action)
	}
}

func newIssueCommentEvent(req IssueCommentCreatedRequest, provider eventpb.EventProvider, deliveryUID string) *eventpb.Event {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED,
		Provider:       provider,
		Time:           timestamppb.New(req.Comment.CreatedAt),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_IssueCommentedPayload{
			IssueCommentedPayload: &eventpb.IssueCommentedPayload{
				UserId:        req.Comment.User.ID,
				IssueNumber:   req.Issue.Number,
				IssueAuthorId: req.Issue.User.ID,
				CommentLength: int32(len(req.Comment.Body)),
				ContainsCode:  containsCode(req.Comment.Body),
			},
		},
	}

	return ev
}

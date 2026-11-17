package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandleIssuesEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) error {
	ev, err := buildIssuesEvent(provider, action, body, deliveryUID)
	if err != nil {
		return err
	}
	return s.saveEvent(context.Background(), ev)
}

func buildIssuesEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) (*eventpb.Event, error) {
	switch action {
	case "opened":
		var req IssueOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, err
		}
		return newIssueOpenedEvent(req, provider, deliveryUID), nil

	case "closed":
		var req IssueClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return nil, err
		}
		return newIssueClosedEvent(req, provider, deliveryUID)

	default:
		return nil, fmt.Errorf("issue action '%s' not handled", action)
	}
}

func newIssueOpenedEvent(req IssueOpenedRequest, provider eventpb.EventProvider, deliveryUID string) *eventpb.Event {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_OPENED,
		Provider:       provider,
		Time:           timestamppb.New(req.Issue.CreatedAt),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_IssueOpenedPayload{
			IssueOpenedPayload: &eventpb.IssueOpenedPayload{
				UserId:      req.Issue.User.ID,
				IssueNumber: req.Issue.Number,
				Title:       req.Issue.Title,
				Labels:      extractLabelsNames(req.Issue.Labels),
			},
		},
	}

	return ev
}

func newIssueClosedEvent(req IssueClosedRequest, provider eventpb.EventProvider, deliveryUID string) (*eventpb.Event, error) {
	closeReason := eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_UNSPECIFIED
	if v := req.Issue.StateReason; v != nil {
		switch *v {
		case "not_planned":
			closeReason = eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_NOT_PLANNED
		case "completed":
			closeReason = eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_COMPLETED
		case "reopened":
			closeReason = eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_REOPENED
		}
	}

	if req.Issue.ClosedAt == nil {
		return nil, fmt.Errorf("invalid IssueClosed payload: issue.closed_at is nil")
	}

	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_CLOSED,
		Provider:       provider,
		Time:           timestamppb.New(*req.Issue.ClosedAt),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_IssueClosedPayload{
			IssueClosedPayload: &eventpb.IssueClosedPayload{
				UserId:        req.Sender.ID,
				IssueAuthorId: req.Issue.User.ID,
				IssueId:       req.Issue.ID,
				IssueNumber:   req.Issue.Number,
				CloseReason:   closeReason,
				Labels:        extractLabelsNames(req.Issue.Labels),
				OpenedAt:      timestamppb.New(req.Issue.CreatedAt),
				CommentsCount: req.Issue.Comments,
				ClosingPrNumber: func() int32 {
					if req.Issue.PullRequest != nil {
						return req.Issue.PullRequest.Number
					}
					return 0
				}(),
			},
		},
	}

	return ev, nil
}

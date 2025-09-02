package service

import (
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandleIssuesEvent(action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req IssueOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.publishIssueOpened(req, deliveryUID)

	case "closed":
		var req IssueClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.publishIssueClosed(req, deliveryUID)

	default:
		return fmt.Errorf("issue action '%s' not handled", action)
	}
}

func (s *Service) publishIssueOpened(req IssueOpenedRequest, deliveryUID string) error {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_ISSUE_OPENED,
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
	metadata := map[string]string{}

	return s.publishEvent(ev, eventpb.EventName_ISSUE_OPENED, TopicGithubIssues, metadata)
}

func (s *Service) publishIssueClosed(req IssueClosedRequest, deliveryUID string) error {

	closeReason := eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_UNSPECIFIED
	if v := req.Issue.StateReason; v != nil {
		switch *v {
		case "not_planned":
			closeReason = eventpb.IssueCloseReason_NOT_PLANNED
		case "completed":
			closeReason = eventpb.IssueCloseReason_COMPLETED
		case "reopened":
			closeReason = eventpb.IssueCloseReason_REOPENED
		}
	}

	if req.Issue.ClosedAt == nil {
		return fmt.Errorf("invalid IssueClosed payload: issue.closed_at is nil")
	}

	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_ISSUE_CLOSED,
		Time:           timestamppb.New(*req.Issue.ClosedAt),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_IssueClosedPayload{
			IssueClosedPayload: &eventpb.IssueClosedPayload{
				UserId:          req.Sender.ID,
				IssueAuthorId:   req.Issue.User.ID,
				IssueId:         req.Issue.ID,
				IssueNumber:     req.Issue.Number,
				CloseReason:     closeReason,
				Labels:          extractLabelsNames(req.Issue.Labels),
				OpenedAt:        timestamppb.New(req.Issue.CreatedAt),
				CommentsCount:   req.Issue.Comments,
				ClosingPrNumber: req.Issue.PullRequest.Number,
			},
		},
	}

	metadata := map[string]string{}

	return s.publishEvent(ev, eventpb.EventName_ISSUE_CLOSED, TopicGithubIssues, metadata)
}

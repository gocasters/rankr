package service

import (
	"encoding/json"
	"fmt"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandleIssueCommentEvent(action string, body []byte, deliveryUID string) error {
	switch action {
	case "created":
		var req IssueCommentCreatedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.publishIssueComment(req, deliveryUID)
	default:
		return fmt.Errorf("issue_comment action '%s' not handled", action)
	}
}

func (s *Service) publishIssueComment(req IssueCommentCreatedRequest, deliveryUID string) error {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED,
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

	metadata := map[string]string{}

	return s.publishEvent(ev, eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED, TopicGithubIssueComment, metadata)
}

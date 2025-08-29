package service

import (
	"encoding/json"
	"fmt"

	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandlePullRequestEvent(action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req PullRequestOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		return s.publishPullRequestOpened(req, deliveryUID)

	case "closed":
		var req PullRequestClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.publishPullRequestClosed(req, deliveryUID)

	default:
		return fmt.Errorf("pull request action '%s' not handled", action)
	}
}

func (s *Service) publishPullRequestOpened(req PullRequestOpenedRequest, deliveryUID string) error {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       req.Sender.ID,
				PrId:         req.PullRequest.ID,
				PrNumber:     req.PullRequest.Number,
				Title:        req.PullRequest.Title,
				BranchName:   req.PullRequest.Head.Ref,
				TargetBranch: req.PullRequest.Base.Ref,
				Labels:       extractLabelsNames(req.PullRequest.Labels),
				Assignees:    extractAssigneesIDs(req.PullRequest.Assignees),
			},
		},
	}

	metadata := map[string]string{}

	return s.publishEvent(ev, eventpb.EventName_PULL_REQUEST_OPENED, TopicGithubPullRequest, metadata)
}

func (s *Service) publishPullRequestClosed(req PullRequestClosedRequest, deliveryUID string) error {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_PULL_REQUEST_CLOSED,
		Time:           timestamppb.Now(),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_PrClosedPayload{
			PrClosedPayload: &eventpb.PullRequestClosedPayload{
				UserId:       req.Sender.ID,
				MergerUserId: req.Sender.ID,
				PrId:         req.PullRequest.ID,
				PrNumber:     req.PullRequest.Number,
				Merged:       req.PullRequest.Merged,
				CloseReason:  determineCloseReason(req.PullRequest.Merged),
				Additions:    req.PullRequest.Additions,
				Deletions:    req.PullRequest.Deletions,
				FilesChanged: req.PullRequest.ChangedFiles,
				CommitsCount: req.PullRequest.Commits,
				Labels:       extractLabelsNames(req.PullRequest.Labels),
				TargetBranch: req.PullRequest.Base.Ref,
				// Documentation-specific fields (zero values if not a doc PR)
				IsDocumentation:    false,
				DocumentationTypes: []string{},
				Assignees:          extractAssigneesIDs(req.PullRequest.Assignees),
			},
		},
	}

	metadata := map[string]string{}

	return s.publishEvent(ev, eventpb.EventName_PULL_REQUEST_CLOSED, TopicGithubPullRequest, metadata)
}

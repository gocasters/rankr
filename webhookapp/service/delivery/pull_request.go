package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandlePullRequestEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) error {
	switch action {
	case "opened":
		var req PullRequestOpenedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		return s.publishPullRequestOpened(req, provider, deliveryUID)

	case "closed":
		var req PullRequestClosedRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return s.publishPullRequestClosed(req, provider, deliveryUID)

	default:
		return fmt.Errorf("pull request action '%s' not handled", action)
	}
}

func (s *Service) publishPullRequestOpened(req PullRequestOpenedRequest, provider eventpb.EventProvider, deliveryUID string) error {
	_ = deliveryUID
	ev := &eventpb.Event{
		Id:             watermill.NewUUID(),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Provider:       provider,
		Time:           timestamppb.New(req.PullRequest.CreatedAt),
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

	ctx := context.Background()
	return s.saveEvent(ctx, ev)
	//return s.publishEvent(ev, eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED, TopicGithubPullRequest, metadata)
}

func (s *Service) publishPullRequestClosed(req PullRequestClosedRequest, provider eventpb.EventProvider, deliveryUID string) error {
	_ = deliveryUID
	t := timestamppb.New(time.Time{}) // it is the "zero time" to be distinguishable
	if req.PullRequest.ClosedAt != nil {
		t = timestamppb.New(*req.PullRequest.ClosedAt)
	}
	ev := &eventpb.Event{
		Id:             watermill.NewUUID(),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
		Provider:       provider,
		Time:           t,
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_PrClosedPayload{
			PrClosedPayload: &eventpb.PullRequestClosedPayload{
				UserId: req.Sender.ID,
				MergerUserId: func() uint64 {
					if req.PullRequest.Merged != nil && *req.PullRequest.Merged && req.PullRequest.MergedBy != nil {
						return req.PullRequest.MergedBy.ID
					}
					return req.Sender.ID
				}(),
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
				Assignees:    extractAssigneesIDs(req.PullRequest.Assignees),
			},
		},
	}

	ctx := context.Background()
	return s.saveEvent(ctx, ev)
	//return s.publishEvent(ev, eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED, TopicGithubPullRequest, metadata)
}

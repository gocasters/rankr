package delivery

import (
	"context"
	"encoding/json"
	"strings"


	"github.com/ThreeDotsLabs/watermill"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandlePushEvent(provider eventpb.EventProvider, body []byte, deliveryUID string) error {
	var req PushRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return err
	}
	return s.publishPush(req, provider, deliveryUID)
}

func (s *Service) publishPush(req PushRequest, provider eventpb.EventProvider, deliveryUID string) error {
	ref := req.Ref
	branchName := strings.TrimPrefix(ref, "refs/heads/")

	commits := req.Commits
	commitInfos := make([]*eventpb.CommitInfo, 0, len(commits))

	for _, commit := range commits {
		commitInfo := &eventpb.CommitInfo{
			AuthorName: commit.Author.Name,
			CommitId:   commit.ID,
			Message:    commit.Message,
			Additions:  int32(len(commit.Added)),
			Deletions:  int32(len(commit.Removed)),
			Modified:   int32(len(commit.Modified)),
		}

		commitInfos = append(commitInfos, commitInfo)
	}

	_ = deliveryUID
	ev := &eventpb.Event{
		Id:        watermill.NewUUID(),
		EventName: eventpb.EventName_EVENT_NAME_PUSHED,
		Provider:  provider,
		//TODO we have no time for when push happened
		Time: func() *timestamppb.Timestamp {
			if req.HeadCommit != nil && req.HeadCommit.Timestamp != "" {
				return timestamppb.New(parseTime(req.HeadCommit.Timestamp))
			}
			return timestamppb.Now()
		}(),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_PushPayload{
			PushPayload: &eventpb.PushPayload{
				UserId:       req.Sender.ID,
				BranchName:   branchName,
				CommitsCount: int32(len(commits)),
				Commits:      commitInfos,
			},
		},
	}

	ctx := context.Background()
	return s.saveEvent(ctx, ev)
	//return s.publishEvent(ev, eventpb.EventName_EVENT_NAME_PUSHED, TopicGithubPush, metadata)
}

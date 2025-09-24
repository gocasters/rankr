package publishevent

import (
	"encoding/json"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

func (s *Service) HandlePushEvent(provider eventpb.EventProvider, body []byte, deliveryUID string) error {
	var req service.PushRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return err
	}
	return s.publishPush(req, provider, deliveryUID)
}

func (s *Service) publishPush(req service.PushRequest, provider eventpb.EventProvider, deliveryUID string) error {
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

	ev := &eventpb.Event{
		Id:        deliveryUID,
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

	metadata := map[string]string{}

	return s.publishEvent(ev, eventpb.EventName_EVENT_NAME_PUSHED, service.TopicGithubPush, metadata)
}

package eventgenerator

import (
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math/rand"
	"time"
)

// EventGenerator generates test events
type EventGenerator struct {
	repoID   uint64
	repoName string
}

func NewEventGenerator(repoID uint64, repoName string) *EventGenerator {
	return &EventGenerator{
		repoID:   repoID,
		repoName: repoName,
	}
}

// GeneratePullRequestOpened creates a PR opened event
func (g *EventGenerator) GeneratePullRequestOpened(userID, prID uint64, prNumber int32) *eventpb.Event {
	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       userID,
				PrId:         prID,
				PrNumber:     prNumber,
				Title:        "Test PR: Add new feature",
				BranchName:   "feature/test-branch",
				TargetBranch: "main",
				Labels:       []string{"enhancement", "test"},
				Assignees:    []uint64{userID},
			},
		},
	}
}

// GeneratePullRequestClosed creates a PR closed event
func (g *EventGenerator) GeneratePullRequestClosed(userID, mergerID, prID uint64, prNumber int32) *eventpb.Event {
	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrClosedPayload{
			PrClosedPayload: &eventpb.PullRequestClosedPayload{
				UserId:       userID,
				MergerUserId: mergerID,
				PrId:         prID,
				PrNumber:     prNumber,
				CloseReason:  eventpb.PrCloseReason_PR_CLOSE_REASON_MERGED,
				Merged:       boolPtr(true),
				Additions:    150,
				Deletions:    50,
				FilesChanged: 5,
				CommitsCount: 3,
				Labels:       []string{"bug", "priority-high"},
				TargetBranch: "main",
				Assignees:    []uint64{userID},
			},
		},
	}
}

// GeneratePullRequestReview creates a PR review event
func (g *EventGenerator) GeneratePullRequestReview(reviewerID, authorID, prID uint64, prNumber int32, state eventpb.ReviewState) *eventpb.Event {
	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrReviewPayload{
			PrReviewPayload: &eventpb.PullRequestReviewSubmittedPayload{
				ReviewerUserId: reviewerID,
				PrAuthorUserId: authorID,
				PrId:           prID,
				PrNumber:       prNumber,
				State:          state,
			},
		},
	}
}

// GenerateIssueOpened creates an issue opened event
func (g *EventGenerator) GenerateIssueOpened(userID, issueID uint64, issueNumber int32) *eventpb.Event {
	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_IssueOpenedPayload{
			IssueOpenedPayload: &eventpb.IssueOpenedPayload{
				UserId:      userID,
				IssueId:     issueID,
				IssueNumber: issueNumber,
				Title:       "Bug: Application crashes on startup",
				Labels:      []string{"bug", "urgent"},
			},
		},
	}
}

// GenerateIssueClosed creates an issue closed event
func (g *EventGenerator) GenerateIssueClosed(userID, authorID, issueID uint64, issueNumber int32) *eventpb.Event {
	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_CLOSED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_IssueClosedPayload{
			IssueClosedPayload: &eventpb.IssueClosedPayload{
				UserId:          userID,
				IssueAuthorId:   authorID,
				IssueId:         issueID,
				IssueNumber:     issueNumber,
				CloseReason:     eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_COMPLETED,
				Labels:          []string{"bug", "fixed"},
				OpenedAt:        timestamppb.New(time.Now().Add(-24 * time.Hour)),
				CommentsCount:   5,
				ClosingPrNumber: 123,
			},
		},
	}
}

// GenerateIssueCommented creates an issue comment event
func (g *EventGenerator) GenerateIssueCommented(userID, authorID, issueID uint64, issueNumber int32) *eventpb.Event {
	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_IssueCommentedPayload{
			IssueCommentedPayload: &eventpb.IssueCommentedPayload{
				UserId:        userID,
				IssueAuthorId: authorID,
				IssueId:       issueID,
				IssueNumber:   issueNumber,
				CommentLength: 250,
				ContainsCode:  true,
			},
		},
	}
}

// GeneratePush creates a push event
func (g *EventGenerator) GeneratePush(userID uint64, branch string, commitsCount int32) *eventpb.Event {
	commits := make([]*eventpb.CommitInfo, commitsCount)
	for i := int32(0); i < commitsCount; i++ {
		commits[i] = &eventpb.CommitInfo{
			AuthorName: "Test User",
			CommitId:   generateCommitID(),
			Message:    "feat: add new feature",
			Additions:  50,
			Deletions:  10,
			Modified:   5,
		}
	}

	return &eventpb.Event{
		Id:             generateEventID(),
		EventName:      eventpb.EventName_EVENT_NAME_PUSHED,
		Time:           timestamppb.Now(),
		RepositoryId:   g.repoID,
		RepositoryName: g.repoName,
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PushPayload{
			PushPayload: &eventpb.PushPayload{
				UserId:       userID,
				BranchName:   branch,
				CommitsCount: commitsCount,
				Commits:      commits,
			},
		},
	}
}

// GenerateRandomEvents generates a mix of random events
func (g *EventGenerator) GenerateRandomEvents(count int, userIDs []uint64) []*eventpb.Event {
	events := make([]*eventpb.Event, 0, count)

	eventTypes := []eventpb.EventName{
		eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
		eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
		eventpb.EventName_EVENT_NAME_ISSUE_OPENED,
		eventpb.EventName_EVENT_NAME_ISSUE_CLOSED,
		eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED,
		eventpb.EventName_EVENT_NAME_PUSHED,
	}

	for i := 0; i < count; i++ {
		userID := userIDs[i%len(userIDs)]
		eventType := eventTypes[i%len(eventTypes)]

		var event *eventpb.Event
		switch eventType {
		case eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED:
			event = g.GeneratePullRequestOpened(userID, uint64(1000+i), int32(100+i))
		case eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED:
			event = g.GeneratePullRequestClosed(userID, userID, uint64(1000+i), int32(100+i))
		case eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED:
			event = g.GeneratePullRequestReview(userID, userIDs[(i+1)%len(userIDs)], uint64(1000+i), int32(100+i), eventpb.ReviewState_REVIEW_STATE_APPROVED)
		case eventpb.EventName_EVENT_NAME_ISSUE_OPENED:
			event = g.GenerateIssueOpened(userID, uint64(2000+i), int32(200+i))
		case eventpb.EventName_EVENT_NAME_ISSUE_CLOSED:
			event = g.GenerateIssueClosed(userID, userID, uint64(2000+i), int32(200+i))
		case eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED:
			event = g.GenerateIssueCommented(userID, userIDs[(i+1)%len(userIDs)], uint64(2000+i), int32(200+i))
		case eventpb.EventName_EVENT_NAME_PUSHED:
			event = g.GeneratePush(userID, "main", 3)
		}

		events = append(events, event)
	}

	return events
}

// Helper functions
func generateEventID() string {
	return uuid.New().String()
	//return fmt.Sprintf("evt_%d_%s", time.Now().UnixNano(), randomString(8))
}

func generateCommitID() string {
	return randomString(40)
}

func randomString(n int) string {
	const letters = "abcdef0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func boolPtr(b bool) *bool {
	return &b
}

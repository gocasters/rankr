package leaderboardscoring

import (
	"fmt"
	"github.com/gocasters/rankr/pkg/timettl"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"time"
)

type EventRequest struct {
	ID             string
	EventName      EventName
	RepositoryID   uint64
	RepositoryName string
	Timestamp      time.Time
	Payload        interface{}
}

func NewEventRequest() *EventRequest {
	return &EventRequest{}
}

type PullRequestOpenedPayload struct {
	UserID       uint64
	PrID         uint64
	PrNumber     int32
	Title        string
	BranchName   string
	TargetBranch string
	Labels       []string
	Assignees    []uint64
}

type PrCloseReason int32

const (
	PrCloseReasonUnspecified PrCloseReason = iota
	PrCloseReasonMerged
	PrCloseReasonClosedWithoutMerge
	PrCloseReasonDraftConverted
)

func (pr PrCloseReason) String() string {
	switch pr {
	case PrCloseReasonMerged:
		return "merged"
	case PrCloseReasonClosedWithoutMerge:
		return "closed_without_merge"
	case PrCloseReasonDraftConverted:
		return "draft_converted"
	default:
		return "unknown"
	}
}

type PullRequestClosedPayload struct {
	UserID       uint64
	MergerUserID uint64
	PrID         uint64
	PrNumber     int32
	CloseReason  PrCloseReason
	Merged       bool
	Additions    int32
	Deletions    int32
	FilesChanged int32
	CommitsCount int32
	Labels       []string
	TargetBranch string
	Assignees    []uint64
}

type PullRequestReviewPayload struct {
	ReviewerUserID uint64
	PrAuthorUserID uint64
	PrID           uint64
	PrNumber       int32
}

type IssueOpenedPayload struct {
	UserID      uint64
	IssueID     uint64
	IssueNumber int32
	Title       string
	Labels      []string
}

type IssueCloseReason int32

const (
	IssueCloseReasonUnspecified IssueCloseReason = iota
	IssueCloseReasonCompleted
	IssueCloseReasonNotPlanned
	IssueCloseReasonReopen
)

type IssueClosedPayload struct {
	UserID          uint64
	IssueAuthorID   uint64
	IssueID         uint64
	IssueNumber     int32
	CloseReason     IssueCloseReason
	Labels          []string
	OpenedAt        time.Time
	CommentsCount   int32
	ClosingPrNumber int32
}

type IssueCommentedPayload struct {
	UserID        uint64
	IssueAuthorID uint64
	IssueID       uint64
	IssueNumber   int32
	CommentLength int32
	ContainsCode  bool
}

type CommitInfo struct {
	AuthorName string
	CommitID   string
	Message    string
	Additions  int32
	Deletions  int32
	Modified   int32
}

type PushPayload struct {
	UserID       uint64
	BranchName   string
	CommitsCount int32
	Commits      []*CommitInfo
}

func (er *EventRequest) ProtobufToEventRequest(eventPB *eventpb.Event) (*EventRequest, error) {
	if eventPB == nil {
		return &EventRequest{}, fmt.Errorf("nil event")
	}

	ts := eventPB.GetTime()
	if ts == nil || !ts.IsValid() {
		return &EventRequest{}, fmt.Errorf("event with ID %s has a missing timestamp", eventPB.Id)
	}

	payload, err := protobufToPayload(eventPB)
	if err != nil {
		return &EventRequest{}, err
	}

	contributionEvent := &EventRequest{
		ID:             eventPB.Id,
		EventName:      EventName(eventPB.GetEventName()),
		RepositoryID:   eventPB.RepositoryId,
		RepositoryName: eventPB.RepositoryName,
		Timestamp:      ts.AsTime().UTC(),
		Payload:        payload,
	}

	return contributionEvent, nil
}

func protobufToPayload(eventPB *eventpb.Event) (interface{}, error) {
	var payload interface{}

	switch eventPB.GetEventName() {
	case eventpb.EventName_PULL_REQUEST_OPENED:
		p := eventPB.GetPrOpenedPayload()
		prPayload := PullRequestOpenedPayload{
			UserID:       p.GetUserId(),
			PrID:         p.GetPrId(),
			PrNumber:     p.GetPrNumber(),
			Title:        p.GetTitle(),
			BranchName:   p.GetBranchName(),
			TargetBranch: p.GetTargetBranch(),
			Labels:       p.GetLabels(),
			Assignees:    p.GetAssignees(),
		}
		payload = prPayload

	case eventpb.EventName_PULL_REQUEST_CLOSED:
		p := eventPB.GetPrClosedPayload()
		prClosedPayload := PullRequestClosedPayload{
			UserID:       p.GetUserId(),
			MergerUserID: p.GetMergerUserId(),
			PrID:         p.GetPrId(),
			PrNumber:     p.GetPrNumber(),
			CloseReason:  PrCloseReason(p.GetCloseReason()),
			Merged:       p.GetMerged(),
			Additions:    p.GetAdditions(),
			Deletions:    p.GetDeletions(),
			FilesChanged: p.GetFilesChanged(),
			CommitsCount: p.GetCommitsCount(),
			Labels:       p.GetLabels(),
			TargetBranch: p.GetTargetBranch(),
			Assignees:    p.GetAssignees(),
		}
		payload = prClosedPayload

	case eventpb.EventName_PULL_REQUEST_REVIEW_SUBMITTED:
		p := eventPB.GetPrReviewPayload()
		prReviewPayload := PullRequestReviewPayload{
			ReviewerUserID: p.GetReviewerUserId(),
			PrAuthorUserID: p.GetPrAuthorUserId(),
			PrID:           p.GetPrId(),
			PrNumber:       p.GetPrNumber(),
		}
		payload = prReviewPayload

	case eventpb.EventName_ISSUE_OPENED:
		p := eventPB.GetIssueOpenedPayload()
		issueOpened := IssueOpenedPayload{
			UserID:      p.GetUserId(),
			IssueID:     p.GetIssueId(),
			IssueNumber: p.GetIssueNumber(),
			Title:       p.GetTitle(),
			Labels:      p.GetLabels(),
		}
		payload = issueOpened

	case eventpb.EventName_ISSUE_CLOSED:
		p := eventPB.GetIssueClosedPayload()
		issueClosed := IssueClosedPayload{
			UserID:          p.GetUserId(),
			IssueAuthorID:   p.GetIssueAuthorId(),
			IssueID:         p.GetIssueId(),
			IssueNumber:     p.GetIssueNumber(),
			CloseReason:     IssueCloseReason(p.GetCloseReason()),
			Labels:          p.GetLabels(),
			OpenedAt:        p.GetOpenedAt().AsTime().UTC(),
			CommentsCount:   p.GetCommentsCount(),
			ClosingPrNumber: p.GetClosingPrNumber(),
		}
		payload = issueClosed

	case eventpb.EventName_ISSUE_COMMENTED:
		p := eventPB.GetIssueCommentedPayload()
		issueCommented := IssueCommentedPayload{
			UserID:        p.GetUserId(),
			IssueAuthorID: p.GetIssueAuthorId(),
			IssueID:       p.GetIssueId(),
			IssueNumber:   p.GetIssueNumber(),
			CommentLength: p.GetCommentLength(),
			ContainsCode:  p.GetContainsCode(),
		}
		payload = issueCommented

	case eventpb.EventName_PUSHED:
		p := eventPB.GetPushPayload()

		commitsPB := p.GetCommits()
		commits := make([]*CommitInfo, 0, len(commitsPB))
		for _, c := range commitsPB {
			commitInfo := &CommitInfo{
				AuthorName: c.GetAuthorName(),
				CommitID:   c.GetCommitId(),
				Message:    c.GetMessage(),
				Additions:  c.GetAdditions(),
				Deletions:  c.GetDeletions(),
				Modified:   c.GetModified(),
			}

			commits = append(commits, commitInfo)
		}

		push := PushPayload{
			UserID:       p.GetUserId(),
			BranchName:   p.GetBranchName(),
			CommitsCount: p.GetCommitsCount(),
			Commits:      commits,
		}
		payload = push

	default:
		return nil, fmt.Errorf(
			"unsupported event name: %s (id=%s)",
			eventPB.EventName.String(),
			eventPB.Id,
		)
	}

	return payload, nil
}

type LeaderboardRow struct {
	Rank   uint64
	UserID string
	Score  uint64
}

type GetLeaderboardResponse struct {
	Timeframe       Timeframe
	ProjectID       *string
	LeaderboardRows []LeaderboardRow
}

type GetLeaderboardRequest struct {
	Timeframe Timeframe
	ProjectID *string
	PageSize  int32
	Offset    int32
}

func (q *GetLeaderboardRequest) BuildKey() string {

	key := "leaderboard"

	if q.ProjectID != nil {
		key += fmt.Sprintf(":%s", *q.ProjectID)
	} else {
		key += ":global"
	}

	key += fmt.Sprintf(":%s", q.Timeframe.String())

	var period string
	switch q.Timeframe {
	case Yearly:
		period = timettl.GetYear()
	case Monthly:
		period = timettl.GetMonth()
	case Weekly:
		period = timettl.GetWeek()
	default:
		period = "unknown"
	}

	key += fmt.Sprintf(":%s", period)

	return key
}

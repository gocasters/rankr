package leaderboardscoring

import (
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/pkg/timettl"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	leaderboardscoringpb "github.com/gocasters/rankr/protobuf/golang/leaderboardscoring/v1"
	"strconv"
	"time"
)

type EventPayload interface {
	EventType() string
}

type EventRequest struct {
	ID             string       `json:"id"`
	UserID         string       `json:"user_id"`
	EventName      string       `json:"event_name"`
	RepositoryID   uint64       `json:"repository_id"`
	RepositoryName string       `json:"repository_name"`
	Timestamp      time.Time    `json:"timestamp"`
	Payload        EventPayload `json:"payload"`
}

func NewEventRequest() *EventRequest {
	return &EventRequest{}
}

func (er *EventRequest) String() string {
	data, _ := json.MarshalIndent(er, "", " ")
	return string(data)
}

func (er *EventRequest) MapProtoEventToEventRequest(eventPB *eventpb.Event) (*EventRequest, error) {
	if eventPB == nil {
		return nil, fmt.Errorf("nil event")
	}

	ts := eventPB.GetTime()
	if ts == nil || !ts.IsValid() {
		return nil, fmt.Errorf("event with ID %s has a missing timestamp", eventPB.Id)
	}

	payload, userID, err := protobufToPayload(eventPB)
	if err != nil {
		return nil, err
	}

	contributionEvent := &EventRequest{
		ID:             eventPB.Id,
		UserID:         userID,
		EventName:      payload.EventType(),
		RepositoryID:   eventPB.RepositoryId,
		RepositoryName: eventPB.RepositoryName,
		Timestamp:      ts.AsTime().UTC(),
		Payload:        payload,
	}

	return contributionEvent, nil
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

func (pr PrCloseReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(pr.String())
}

type IssueCloseReason int32

const (
	IssueCloseReasonUnspecified IssueCloseReason = iota
	IssueCloseReasonCompleted
	IssueCloseReasonNotPlanned
	IssueCloseReasonReopen
)

func (ic IssueCloseReason) String() string {
	switch ic {
	case IssueCloseReasonCompleted:
		return "completed"
	case IssueCloseReasonNotPlanned:
		return "not_planned"
	case IssueCloseReasonReopen:
		return "reopen"
	default:
		return "unknown"
	}
}

func (ic IssueCloseReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(ic.String())
}

type PullRequestOpenedPayload struct {
	UserID       uint64   `json:"user_id"`
	PrID         uint64   `json:"pr_id"`
	PrNumber     int32    `json:"pr_number"`
	Title        string   `json:"title"`
	BranchName   string   `json:"branch_name"`
	TargetBranch string   `json:"target_branch"`
	Labels       []string `json:"labels"`
	Assignees    []uint64 `json:"assignees"`
}

func (p PullRequestOpenedPayload) EventType() string {
	return PullRequestOpened.String()
}

type PullRequestClosedPayload struct {
	UserID       uint64        `json:"user_id"`
	MergerUserID uint64        `json:"merger_user_id"`
	PrID         uint64        `json:"pr_id"`
	PrNumber     int32         `json:"pr_number"`
	CloseReason  PrCloseReason `json:"close_reason"`
	Merged       bool          `json:"merged"`
	Additions    int32         `json:"additions"`
	Deletions    int32         `json:"deletions"`
	FilesChanged int32         `json:"files_changed"`
	CommitsCount int32         `json:"commits_count"`
	Labels       []string      `json:"labels"`
	TargetBranch string        `json:"target_branch"`
	Assignees    []uint64      `json:"assignees"`
}

func (p PullRequestClosedPayload) EventType() string {
	return PullRequestClosed.String()
}

type PullRequestReviewPayload struct {
	ReviewerUserID uint64 `json:"reviewer_user_id"`
	PrAuthorUserID uint64 `json:"pr_author_user_id"`
	PrID           uint64 `json:"pr_id"`
	PrNumber       int32  `json:"pr_number"`
}

func (p PullRequestReviewPayload) EventType() string {
	return PullRequestReview.String()
}

type IssueOpenedPayload struct {
	UserID      uint64   `json:"user_id"`
	IssueID     uint64   `json:"issue_id"`
	IssueNumber int32    `json:"issue_number"`
	Title       string   `json:"title"`
	Labels      []string `json:"labels"`
}

func (i IssueOpenedPayload) EventType() string {
	return IssueOpened.String()
}

type IssueClosedPayload struct {
	UserID          uint64           `json:"user_id"`
	IssueAuthorID   uint64           `json:"issue_author_id"`
	IssueID         uint64           `json:"issue_id"`
	IssueNumber     int32            `json:"issue_number"`
	CloseReason     IssueCloseReason `json:"close_reason"`
	Labels          []string         `json:"labels"`
	OpenedAt        time.Time        `json:"opened_at"`
	CommentsCount   int32            `json:"comments_count"`
	ClosingPrNumber int32            `json:"closing_pr_number"`
}

func (i IssueClosedPayload) EventType() string {
	return IssueClosed.String()
}

type IssueCommentedPayload struct {
	UserID        uint64 `json:"user_id"`
	IssueAuthorID uint64 `json:"issue_author_id"`
	IssueID       uint64 `json:"issue_id"`
	IssueNumber   int32  `json:"issue_number"`
	CommentLength int32  `json:"comment_length"`
	ContainsCode  bool   `json:"contains_code"`
}

func (i IssueCommentedPayload) EventType() string {
	return IssueComment.String()
}

type PushPayload struct {
	UserID       uint64        `json:"user_id"`
	BranchName   string        `json:"branch_name"`
	CommitsCount int32         `json:"commits_count"`
	Commits      []*CommitInfo `json:"commits"`
}

func (p PushPayload) EventType() string {
	return CommitPush.String()
}

type CommitInfo struct {
	AuthorName string `json:"author_name"`
	CommitID   string `json:"commit_id"`
	Message    string `json:"message"`
	Additions  int32  `json:"additions"`
	Deletions  int32  `json:"deletions"`
	Modified   int32  `json:"modified"`
}

func protobufToPayload(eventPB *eventpb.Event) (EventPayload, string, error) {
	var payload EventPayload
	var userID string

	switch eventPB.Payload.(type) {
	case *eventpb.Event_PrOpenedPayload:
		p := eventPB.GetPrOpenedPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing pr_opened payload (id=%s)", eventPB.Id)
		}
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
		userID = strconv.FormatUint(prPayload.UserID, 10)

	case *eventpb.Event_PrClosedPayload:
		p := eventPB.GetPrClosedPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing pr_closed payload (id=%s)", eventPB.Id)
		}
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
		userID = strconv.FormatUint(prClosedPayload.UserID, 10)

	case *eventpb.Event_PrReviewPayload:
		p := eventPB.GetPrReviewPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing pr_review payload (id=%s)", eventPB.Id)
		}
		prReviewPayload := PullRequestReviewPayload{
			ReviewerUserID: p.GetReviewerUserId(),
			PrAuthorUserID: p.GetPrAuthorUserId(),
			PrID:           p.GetPrId(),
			PrNumber:       p.GetPrNumber(),
		}
		payload = prReviewPayload
		userID = strconv.FormatUint(prReviewPayload.ReviewerUserID, 10)

	case *eventpb.Event_IssueOpenedPayload:
		p := eventPB.GetIssueOpenedPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing issue_opened payload (id=%s)", eventPB.Id)
		}
		issueOpened := IssueOpenedPayload{
			UserID:      p.GetUserId(),
			IssueID:     p.GetIssueId(),
			IssueNumber: p.GetIssueNumber(),
			Title:       p.GetTitle(),
			Labels:      p.GetLabels(),
		}
		payload = issueOpened
		userID = strconv.FormatUint(issueOpened.UserID, 10)

	case *eventpb.Event_IssueClosedPayload:
		p := eventPB.GetIssueClosedPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing issue_closed payload (id=%s)", eventPB.Id)
		}
		openedAtTS := p.GetOpenedAt()
		if openedAtTS == nil || !openedAtTS.IsValid() {
			return nil, "", fmt.Errorf("missing or invalid issue_closed.opened_at (id=%s)", eventPB.Id)
		}
		issueClosed := IssueClosedPayload{
			UserID:          p.GetUserId(),
			IssueAuthorID:   p.GetIssueAuthorId(),
			IssueID:         p.GetIssueId(),
			IssueNumber:     p.GetIssueNumber(),
			CloseReason:     IssueCloseReason(p.GetCloseReason()),
			Labels:          p.GetLabels(),
			OpenedAt:        openedAtTS.AsTime().UTC(),
			CommentsCount:   p.GetCommentsCount(),
			ClosingPrNumber: p.GetClosingPrNumber(),
		}
		payload = issueClosed
		userID = strconv.FormatUint(issueClosed.UserID, 10)

	case *eventpb.Event_IssueCommentedPayload:
		p := eventPB.GetIssueCommentedPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing issue_commented payload (id=%s)", eventPB.Id)
		}
		issueCommented := IssueCommentedPayload{
			UserID:        p.GetUserId(),
			IssueAuthorID: p.GetIssueAuthorId(),
			IssueID:       p.GetIssueId(),
			IssueNumber:   p.GetIssueNumber(),
			CommentLength: p.GetCommentLength(),
			ContainsCode:  p.GetContainsCode(),
		}
		payload = issueCommented
		userID = strconv.FormatUint(issueCommented.UserID, 10)

	case *eventpb.Event_PushPayload:
		p := eventPB.GetPushPayload()
		if p == nil {
			return nil, "", fmt.Errorf("missing pushed payload (id=%s)", eventPB.Id)
		}
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
		userID = strconv.FormatUint(push.UserID, 10)

	default:
		return nil, "",
			fmt.Errorf(
				"unsupported event name: %s (id=%s)",
				eventPB.EventName.String(),
				eventPB.Id,
			)
	}

	return payload, userID, nil
}

type LeaderboardRow struct {
	Rank   int64
	UserID string
	Score  int64
}

type GetLeaderboardResponse struct {
	Timeframe       string
	ProjectID       *string
	LeaderboardRows []LeaderboardRow
}

func ToProtoTimeframe(tf string) leaderboardscoringpb.Timeframe {
	switch tf {
	case AllTime.String():
		return leaderboardscoringpb.Timeframe_TIMEFRAME_ALL_TIME
	case Yearly.String():
		return leaderboardscoringpb.Timeframe_TIMEFRAME_YEARLY
	case Monthly.String():
		return leaderboardscoringpb.Timeframe_TIMEFRAME_MONTHLY
	case Weekly.String():
		return leaderboardscoringpb.Timeframe_TIMEFRAME_WEEKLY
	default:
		return leaderboardscoringpb.Timeframe_TIMEFRAME_UNSPECIFIED
	}
}

func FromProtoTimeframe(tf leaderboardscoringpb.Timeframe) string {
	switch tf {
	case leaderboardscoringpb.Timeframe_TIMEFRAME_ALL_TIME:
		return AllTime.String()
	case leaderboardscoringpb.Timeframe_TIMEFRAME_YEARLY:
		return Yearly.String()
	case leaderboardscoringpb.Timeframe_TIMEFRAME_MONTHLY:
		return Monthly.String()
	case leaderboardscoringpb.Timeframe_TIMEFRAME_WEEKLY:
		return Weekly.String()
	default:
		return TimeframeUnspecified.String()
	}
}

type GetLeaderboardRequest struct {
	Timeframe string
	ProjectID *string
	PageSize  int32
	Offset    int32
}

// leaderboard:global:all_time
// leaderboard:1001:all_time
func (q *GetLeaderboardRequest) BuildKey() string {

	key := "leaderboard"

	if q.ProjectID != nil {
		key += fmt.Sprintf(":%s", *q.ProjectID)
	} else {
		key += ":global"
	}

	key += fmt.Sprintf(":%s", q.Timeframe)

	var period string
	switch q.Timeframe {
	case Yearly.String():
		period = timettl.GetYear()
	case Monthly.String():
		period = timettl.GetMonth()
	case Weekly.String():
		period = timettl.GetWeek()
	case AllTime.String():
		return key
	default:
		period = "unknown"
	}

	key += fmt.Sprintf(":%s", period)

	return key
}

package leaderboardscoring

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"time"
)

type EventName string

const (
	PullRequestOpened EventName = "pull_request_opened"
	PullRequestClosed EventName = "pull_request_closed"
	PullRequestReview EventName = "pull_request_review"

	IssueOpened  EventName = "issue_opened"
	IssueClosed  EventName = "issue_closed"
	IssueComment EventName = "issue_comment"

	CommitPush EventName = "commit_push"
)

func (e EventName) Validate() error {
	return validation.Validate(string(e),
		validation.Required,
		validation.In(
			string(PullRequestOpened),
			string(PullRequestClosed),
			string(PullRequestReview),
			string(IssueOpened),
			string(IssueClosed),
			string(IssueComment),
			string(CommitPush),
		),
	)
}

func (e EventName) String() string {
	switch e {
	case PullRequestOpened:
		return "pull_request_opened"
	case PullRequestClosed:
		return "pull_request_closed"
	case PullRequestReview:
		return "pull_request_review"
	case IssueOpened:
		return "issue_opened"
	case IssueClosed:
		return "issue_closed"
	case IssueComment:
		return "issue_comment"
	case CommitPush:
		return "commit_push"
	default:
		return "unknown"
	}
}

type Event struct {
	ID             string
	EventName      EventName
	RepositoryID   uint64
	RepositoryName string
	ContributorID  int
	Score          int64
	Timestamp      time.Time // UTC
}

type Timeframe int

const (
	TimeframeUnspecified Timeframe = iota
	AllTime
	Yearly
	Monthly
	Weekly
)

var Timeframes = []Timeframe{
	AllTime,
	Yearly,
	Monthly,
	Weekly,
}

func (tf Timeframe) String() string {
	switch tf {
	case AllTime:
		return "all_time"
	case Yearly:
		return "yearly"
	case Monthly:
		return "monthly"
	case Weekly:
		return "weekly"
	default:
		return "unknown"
	}
}

type UpsertScore struct {
	Keys   []string
	Score  int64
	UserID string
}

type LeaderboardQuery struct {
	Key   string
	Start int64
	Stop  int64
}

type LeaderboardEntry struct {
	Rank   uint64
	UserID string
	Score  int64
}

type LeaderboardQueryResult struct {
	LeaderboardRows []LeaderboardEntry
}

type ProcessedScoreEvent struct {
	ID        uint64    `json:"id"`
	UserID    string    `json:"user_id"`
	EventName EventName `json:"event_name"`
	Score     int64     `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

type UserTotalScore struct {
	ID                uint64
	UserID            string
	TotalScore        int64
	SnapshotTimestamp time.Time
}

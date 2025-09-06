package leaderboardscoring

import (
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

var EventNames = []EventName{
	PullRequestOpened,
	PullRequestClosed,
	PullRequestReview,
	IssueOpened,
	IssueClosed,
	IssueComment,
	CommitPush,
}

type Event struct {
	ID             string
	EventName      EventName
	RepositoryID   uint64
	RepositoryName string
	ContributorID  int
	Score          int
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
	Score  uint8
	UserID string
}

package leaderboardscoring

import (
	"time"
)

type EventType string

const (
	PullRequestOpened EventType = "pull_request_opened"
	PullRequestClosed EventType = "pull_request_closed"
	PullRequestReview EventType = "pull_request_review"

	IssueOpened  EventType = "issue_opened"
	IssueClosed  EventType = "issue_closed"
	IssueComment EventType = "issue_comment"

	CommitPush EventType = "commit_push"
)

var EventTypes = []EventType{
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
	EventName      EventType
	RepositoryID   uint64
	RepositoryName string
	ContributorID  string
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

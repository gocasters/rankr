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

type EventRequest struct {
	ID             string
	EventName      EventType
	RepositoryID   uint64
	RepositoryName string
	ContributorID  uint64
	Timestamp      time.Time
}

type GetLeaderboardRequest struct {
}

type GetLeaderboardResponse struct {
}

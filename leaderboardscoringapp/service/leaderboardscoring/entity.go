package leaderboardscoring

import "time"

type ContributionType string

const (
	ContributionCommit      ContributionType = "commit"
	ContributionReview      ContributionType = "review"
	ContributionIssueClosed ContributionType = "issue_closed"
)

type ContributionEvent struct {
	ID              string
	Type            ContributionType
	EventName       string
	RepositoryID    uint64
	RepositoryName  string
	ContributorID   string
	ScoreValue      int
	SourceReference string
	Timestamp       time.Time // UTC
}

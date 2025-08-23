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
	UserID          string
	ProjectID       string
	Type            ContributionType
	ScoreValue      int
	SourceReference string
	Timestamp       time.Time // UTC
}

package leaderboardscoring

import "time"

type ContributionEvent struct {
	ID              string
	UserID          string
	ProjectID       string
	Type            string // "commit", "review", "issue_closed", ...
	ScoreValue      int
	SourceReference string
	Timestamp       time.Time
}

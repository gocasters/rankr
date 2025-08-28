package leaderboardscoring

import "time"

type EventRequest struct {
	ID              string
	EventName       string
	RepositoryID    uint64
	RepositoryName  string
	SourceReference string
	ContributorID   string
	Timestamp       time.Time
}

type GetLeaderboardRequest struct {
}

type GetLeaderboardResponse struct {
}

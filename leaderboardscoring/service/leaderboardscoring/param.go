package leaderboardscoring

import "time"

type EventRequest struct {
	ID              string
	UserID          string
	ProjectID       string
	Type            string
	ScoreValue      int
	SourceReference string
	Timestamp       time.Time
}

type GetLeaderboardRequest struct {
}

type GetLeaderboardResponse struct {
}

package leaderboardscoring

import "time"

type Event struct {
	ID             string
	EventName      string
	RepositoryID   uint64
	RepositoryName string
	ContributorID  string
	Score          int
	Timestamp      time.Time // UTC
}

package leaderboardscoring

import (
	"fmt"
	"github.com/gocasters/rankr/pkg/timettl"
	"time"
)

type EventRequest struct {
	ID             string
	EventName      EventType
	RepositoryID   uint64
	RepositoryName string
	ContributorID  uint64
	Timestamp      time.Time
}

type LeaderboardRow struct {
	Rank   uint64
	UserID string
	Score  uint64
}

type GetLeaderboardResponse struct {
	Timeframe       Timeframe
	ProjectID       *string
	LeaderboardRows []LeaderboardRow
}

type GetLeaderboardRequest struct {
	Timeframe Timeframe
	ProjectID *string
	PageSize  int32
	Offset    int32
}

func (q *GetLeaderboardRequest) BuildKey() string {

	key := "leaderboard"

	if q.ProjectID != nil {
		key += fmt.Sprintf(":%s", *q.ProjectID)
	} else {
		key += ":global"
	}

	key += fmt.Sprintf(":%s", q.Timeframe.String())

	var period string
	switch q.Timeframe {
	case Yearly:
		period = timettl.GetYear()
	case Monthly:
		period = timettl.GetMonth()
	case Weekly:
		period = timettl.GetWeek()
	default:
		period = "unknown"
	}

	key += fmt.Sprintf(":%s", period)

	return key
}

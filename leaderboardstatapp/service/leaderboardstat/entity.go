package leaderboardstat

import (
	types "github.com/gocasters/rankr/type"
	"time"
)

type DailyContributorScore struct {
	ID            types.ID
	ContributorID types.ID
	UserID        string  //TODO - data type
	DailyScore    float64 //TODO - data type
	Rank          uint64
	Timeframe     string
	CalculatedAt  time.Time
}

type ScoreRecord struct {
	ID            types.ID  `db:"id"`
	ContributorID types.ID  `db:"contributor_id"`
	ProjectID     types.ID  `db:"project_id"`
	Activity      string    `db:"activity"`
	Score         float64   `db:"score"`
	EarnedAt      time.Time `db:"created_at"`
}

type ScoreEntry struct {
	Activity string    `koanf:"activity"`
	Score    float64   `koanf:"score"`
	EarnedAt time.Time `koanf:"earned_at"`
}

type ContributorStats struct {
	ContributorID types.ID                  `koanf:"contributor_id"`
	GlobalRank    uint                      `koanf:"global_rank"`
	TotalScore    float64                   `koanf:"total_score"`
	ProjectsScore map[types.ID]float64      `koanf:"project_score"`
	ScoreHistory  map[types.ID][]ScoreEntry `koanf:"score_history"`
}

type ContributorTotalStats struct {
	ContributorID types.ID             `koanf:"contributor_id"`
	GlobalRank    uint                 `koanf:"global_rank"`
	TotalScore    float64              `koanf:"total_score"`
	ProjectsScore map[types.ID]float64 `koanf:"project_score"`
}

type UserScore struct {
	UserID types.ID
	Score  float64
}

type ProjectScore struct {
	UserID    types.ID
	ProjectID types.ID
	Score     float64
}

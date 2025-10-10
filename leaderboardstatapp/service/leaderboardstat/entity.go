package leaderboardstat

import (
	types "github.com/gocasters/rankr/type"
	"time"
)

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

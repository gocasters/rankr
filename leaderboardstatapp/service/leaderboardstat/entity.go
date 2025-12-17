package leaderboardstat

import (
	types "github.com/gocasters/rankr/type"
	"time"
)

type UserProjectScore struct {
	ID            types.ID `json:"id"`
	ContributorID types.ID `json:"contributor_id"`
	ProjectID     types.ID `json:"project_id"`
	Score         float64  `json:"score"`
	Timeframe     string   `json:"timeframe"`
	TimeValue     string   `json:"time-value"`
}

type DailyContributorScore struct {
	ID            types.ID
	ContributorID types.ID
	UserID        string
	Score         float64 //TODO - data type
	ProjectID     types.ID
	Rank          int64
	Timeframe     string
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

type ProjectScoreList struct {
	ProjectID   types.ID    `koanf:"project_id"`
	UsersScore  []UserScore `koanf:"user_scores"`
	Total       uint64      `koanf:"total"`
	Page        int32       `koanf:"page"`
	PageSize    int32       `koanf:"page_size"`
	TotalPages  int32       `koanf:"total_pages"`
	LastUpdated *time.Time  `koanf:"last_updated"`
}

type UserScore struct {
	ContributorID types.ID `koanf:"contributor_id"`
	Score         float64  `koanf:"score"`
	Rank          uint64   `koanf:"rank"`
}

type UserScoreEntry struct {
	UserID int     `json:"user_id"`
	Score  float64 `json:"score"`
}

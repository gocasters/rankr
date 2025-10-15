package leaderboardstat

import types "github.com/gocasters/rankr/type"

type ContributorStatsRequest struct {
	ContributorID types.ID
}
type ContributorStatsResponse struct {
	ContributorID types.ID                  `koanf:"contributor_id"`
	GlobalRank    int                       `koanf:"global_rank"`
	TotalScore    float64                   `koanf:"total_score"`
	ProjectsScore map[types.ID]float64      `koanf:"project_score"`
	ScoreHistory  map[types.ID][]ScoreEntry `koanf:"score_history"`
}

type ScoresListResponse struct{}

type CategoryList int

const (
	global CategoryList = iota
	cdp
	eebi
	mapserver
	beehive
	rankr
)

type LeaderboardFilterRequest struct {
	Category  CategoryList `koanf:"category"`
	Timeframe string       `koanf:"timeframe"`
	Page      int          `koanf:"page"`
	PageSize  int          `koanf:"page_size"`
}

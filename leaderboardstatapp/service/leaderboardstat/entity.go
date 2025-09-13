package leaderboardstat

import types "github.com/gocasters/rankr/type"

type ScoreboardItem struct {
	Rank          int `koanf:"rank"`
	ContributorID int `koanf:"contributor_id"`
	Score         int `koanf:"score"`
}

type LeaderboardEntry struct {
	ContributorID string  `koanf:"contributor_id"`
	Score         float64 `koanf:"score"`
}

type ContributorStat struct {
	ContributorID types.ID       `koanf:"contributor_id"`
	GlobalRank    int            `koanf:"global_rank"`
	TotalScore    float64        `koanf:"total_score"`
	ProjectsScore map[string]int `koanf:"project_score"`
	ScoreHistory  map[string]int `koanf:"score_history"`
}

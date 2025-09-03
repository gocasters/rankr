package leaderboardstat

type ScoreboardItem struct {
	Rank          int `koanf:"rank"`
	ContributorID int `koanf:"contributor_id"`
	Score         int `koanf:"score"`
}

type LeaderboardEntry struct {
	ContributorID string `koanf:"contributor_id"`
	Score         int    `koanf:"score"`
}

type ContributorStat struct {
	ContributorID int            `koanf:"contributor_id"`
	GlobalRank    int            `koanf:"global_rank"`
	TotalScore    int            `koanf:"total_score"`
	ProjectScore  map[string]int `koanf:"project_score"`
	ScoreHistory  map[string]int `koanf:"score_history"`
}

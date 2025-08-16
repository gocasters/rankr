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

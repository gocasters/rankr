package leaderboardstat

type ScoreboardResponse struct {
	Entries []ScoreboardItem
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

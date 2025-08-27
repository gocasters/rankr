package leaderboardstat

import (
	"context"
)

type Repository interface {
	GetLeaderboardByFilters(ctx context.Context, page int, pageSize int, list CategoryList, timeframe string) ([]ScoreboardItem, error)
	CreateNewScoreList(ctx context.Context, redisKey string, entries []LeaderboardEntry) error
}

type Service struct {
	repository Repository
}

// GetScoreboard for public leaderboard
func (s *Service) GetLeaderboardByFilters(ctx context.Context, req LeaderboardFilterRequest) (ScoreboardResponse, error) {
	page := req.Page
	pageSize := req.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	category := req.Category
	// TODO - timeframe validation
	timeframe := req.Timeframe

	scoreboard, err := s.repository.GetLeaderboardByFilters(ctx, page, pageSize, category, timeframe)
	if err != nil {
		return ScoreboardResponse{}, err
	}

	return ScoreboardResponse{Entries: scoreboard}, nil
}

// GetContributorScores for user profile
func GetContributorScores(contributorID int, project string) ScoresListResponse {

	return ScoresListResponse{}
}

func (s *Service) CreateNewScoreList(ctx context.Context, redisKey string, entries []LeaderboardEntry) error {
	err := s.repository.CreateNewScoreList(ctx, redisKey, entries)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetContributorStats(ctx context.Context, contributorID int) (ContributorStat, error) {
	// TODO - implement functions and calc contributions stats related to this contributor

	stats := ContributorStat{
		ContributorID: contributorID,
		GlobalRank:    0,
		TotalScore:    0,
		ProjectScore:  map[string]int{},
		ScoreHistory:  map[string]int{},
	}
	return stats, nil
}

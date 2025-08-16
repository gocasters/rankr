package leaderboardstat

import (
	"context"
)

type Repository interface {
	GetScoreboardByFilters(ctx context.Context, page int, pageSize int, list CategoryList, timeframe string) ([]ScoreboardItem, error)
	CreateNewScoreList(ctx context.Context, redisKey string, entries []LeaderboardEntry) error
}

type Service struct {
	repository Repository
}

// GetScoreboard for public leaderboard
func (s *Service) GetScoreboardByFilters(ctx context.Context, req ScoreboardFilterRequest) (ScoreboardResponse, error) {
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

	scoreboard, err := s.repository.GetScoreboardByFilters(ctx, page, pageSize, category, timeframe)
	if err != nil {
		// TODO - create Error object
		return ScoreboardResponse{}, err
	}

	return ScoreboardResponse{Entries: scoreboard}, nil
}

// GetContributorScores for user profile
func GetContributorScores(contributorID int, project string) ScoresListResponse {
	return ScoresListResponse{}
}

func (s *Service) CreateNewScoreList(ctx context.Context, redisKey string, entries []LeaderboardEntry) error {
	//ctx := context.Background()
	err := s.repository.CreateNewScoreList(ctx, redisKey, entries)
	if err != nil {
		return err
	}
	return nil
}

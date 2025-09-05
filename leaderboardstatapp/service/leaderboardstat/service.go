package leaderboardstat

import (
	"context"
	"log/slog"
)

type Repository interface {
	//CreateNewScoreList(ctx context.Context, redisKey string, entries []LeaderboardEntry) error
}

type Service struct {
	repository Repository
	validator  Validator
	logger     *slog.Logger
}

func NewService(repo Repository, validator Validator, logger *slog.Logger) Service {
	return Service{
		repository: repo,
		validator:  validator,
		logger:     logger,
	}
}

// GetContributorScores for user profile
func GetContributorScores(contributorID int, project string) ScoresListResponse {

	return ScoresListResponse{}
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

package leaderboardstat

import (
	"context"
	types "github.com/gocasters/rankr/type"
)

type Repository interface {
	GetContributorTotalScore(ctx context.Context, ID types.ID) (float64, error)
}

type Service struct {
	repository Repository
	validator  Validator
}

func NewService(repo Repository, validator Validator) Service {
	return Service{
		repository: repo,
		validator:  validator,
	}
}

// GetContributorScores for user profile
func GetContributorScores(contributorID int, project string) ScoresListResponse {

	return ScoresListResponse{}
}

func (s *Service) GetContributorTotalStats(ctx context.Context, contributorID types.ID) (ContributorStat, error) {
	// TODO - implement functions and calc contributions stats related to this contributor
	scores, _ := s.repository.GetContributorTotalScore(ctx, contributorID)

	stats := ContributorStat{
		ContributorID: contributorID,
		GlobalRank:    1,
		TotalScore:    scores,
		ProjectsScore: map[string]int{},
		ScoreHistory:  map[string]int{},
	}
	return stats, nil
}

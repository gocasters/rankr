package leaderboardstat

import (
	"context"
	types "github.com/gocasters/rankr/type"
)

type Repository interface {
	GetContributorTotalScore(ctx context.Context, ID types.ID) (float64, error)
	GetContributorTotalRank(ctx context.Context, ID types.ID) (uint, error)
	GetContributorProjectScores(ctx context.Context, ID types.ID) (map[types.ID]float64, error)
	GetContributorScoreHistory(ctx context.Context, ID types.ID) ([]ScoreRecord, error)
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

func GetContributorScores(contributorID int, project string) ScoresListResponse {

	return ScoresListResponse{}
}

func (s *Service) buildScoreHistory(records []ScoreRecord) map[types.ID][]ScoreEntry {

	history := make(map[types.ID][]ScoreEntry)
	for _, record := range records {
		entry := ScoreEntry{
			Activity: record.Activity,
			Score:    record.Score,
			EarnedAt: record.EarnedAt,
		}
		history[record.ProjectID] = append(history[record.ProjectID], entry)
	}

	return history
}

func (s *Service) GetContributorStats(ctx context.Context, contributorID types.ID) (ContributorStats, error) {
	totalScore, err := s.repository.GetContributorTotalScore(ctx, contributorID)
	if err != nil {
		return ContributorStats{}, err
	}
	globalRank, err := s.repository.GetContributorTotalRank(ctx, contributorID)
	if err != nil {
		return ContributorStats{}, err
	}
	projectsScore, err := s.repository.GetContributorProjectScores(ctx, contributorID)
	if err != nil {
		return ContributorStats{}, err
	}
	scoreRecords, err := s.repository.GetContributorScoreHistory(ctx, contributorID)
	if err != nil {
		return ContributorStats{}, err
	}

	scoreHistory := s.buildScoreHistory(scoreRecords)

	stats := ContributorStats{
		ContributorID: contributorID,
		GlobalRank:    globalRank,
		TotalScore:    totalScore,
		ProjectsScore: projectsScore,
		ScoreHistory:  scoreHistory,
	}
	return stats, nil
}

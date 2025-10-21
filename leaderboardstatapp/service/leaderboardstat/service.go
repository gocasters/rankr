package leaderboardstat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/pkg/cachemanager"
	leaderboardscoringpb "github.com/gocasters/rankr/protobuf/golang/leaderboardscoring/v1"
	types "github.com/gocasters/rankr/type"
	"time"
)

type LeaderboardScoringRPC interface {
	GetContributorScores(ctx context.Context, contributorID types.ID) (*leaderboardscoringpb.ContributorScoresResponse, error)
}

type Repository interface {
	GetContributorTotalScore(ctx context.Context, ID types.ID) (float64, error)
	GetContributorTotalRank(ctx context.Context, ID types.ID) (uint, error)
	GetContributorProjectScores(ctx context.Context, ID types.ID) (map[types.ID]float64, error)
	GetContributorScoreHistory(ctx context.Context, ID types.ID) ([]ScoreRecord, error)
}

type Service struct {
	repository            Repository
	validator             Validator
	cacheManager          cachemanager.CacheManager
	leaderboardScoringRPC LeaderboardScoringRPC
}

func NewService(repo Repository, validator Validator, cacheManger cachemanager.CacheManager, rpc LeaderboardScoringRPC) Service {
	return Service{
		repository:            repo,
		validator:             validator,
		cacheManager:          cacheManger,
		leaderboardScoringRPC: rpc,
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

func (s *Service) GetContributorTotalStats(ctx context.Context, contributorID types.ID) (ContributorTotalStats, error) {
	cacheKey := fmt.Sprintf("contributor:%d:total_stats", contributorID)
	cached, err := s.cacheManager.Get(ctx, cacheKey)
	if err == nil {
		var stats ContributorTotalStats
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return stats, nil
		}
	}

	var stats ContributorTotalStats
	if s.leaderboardScoringRPC != nil {
		rpcResponse, err := s.leaderboardScoringRPC.GetContributorScores(ctx, contributorID)
		if err != nil {
			return ContributorTotalStats{}, fmt.Errorf("RPC call failed: %v", err)
		}
		stats = ContributorTotalStats{
			ContributorID: contributorID,
			GlobalRank:    uint(rpcResponse.GetGlobalRank()),
			TotalScore:    rpcResponse.GetTotalScore(),
			ProjectsScore: convertProtoProjectsScore(rpcResponse.GetProjectScores()),
		}
	} else {
		// Fallback to repository
		totalScore, err := s.repository.GetContributorTotalScore(ctx, contributorID)
		if err != nil {
			return ContributorTotalStats{}, err
		}
		globalRank, err := s.repository.GetContributorTotalRank(ctx, contributorID)
		if err != nil {
			return ContributorTotalStats{}, err
		}
		projectsScore, err := s.repository.GetContributorProjectScores(ctx, contributorID)
		if err != nil {
			return ContributorTotalStats{}, err
		}
		stats = ContributorTotalStats{
			ContributorID: contributorID,
			GlobalRank:    globalRank,
			TotalScore:    totalScore,
			ProjectsScore: projectsScore,
		}
	}

	if b, mErr := json.Marshal(stats); mErr == nil {
		_ = s.cacheManager.Set(ctx, cacheKey, string(b), 5*time.Minute)
	}

	return stats, nil
}

func convertProtoProjectsScore(protoScores map[uint64]float64) map[types.ID]float64 {
	projectsScore := make(map[types.ID]float64)
	for projectID, score := range protoScores {
		projectsScore[types.ID(projectID)] = score
	}
	return projectsScore
}

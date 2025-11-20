package leaderboardstat

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/leaderboardscoring"
	lbscoring "github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
	"log/slog"
	"sort"
	"time"
)

type LeaderboardScoringRPC interface {
	//GetContributorScores(ctx context.Context, contributorID types.ID) (*leaderboardscoringpb.ContributorScoresResponse, error)
}

type Repository interface {
	GetContributorTotalScore(ctx context.Context, ID types.ID) (float64, error)
	GetContributorTotalRank(ctx context.Context, ID types.ID) (uint, error)
	GetContributorProjectScores(ctx context.Context, ID types.ID) (map[types.ID]float64, error)
	GetContributorScoreHistory(ctx context.Context, ID types.ID) ([]ScoreRecord, error)
	StoreDailyContributorScores(ctx context.Context, scores []DailyContributorScore) error

	GetPendingDailyScores(ctx context.Context) ([]DailyContributorScore, error)
	UpdateUserProjectScores(ctx context.Context, userProjectScores []UserProjectScore) error
	UpdateGlobalScores(ctx context.Context, userProjectScores []UserProjectScore) error
	MarkDailyScoresAsProcessed(ctx context.Context, scoreIDs []types.ID) error
}

type Service struct {
	repository            Repository
	validator             Validator
	cacheManager          cachemanager.CacheManager
	leaderboardScoringRPC LeaderboardScoringRPC
	lbScoringClient       *leaderboardscoring.Client
}

func NewService(repo Repository, validator Validator, cacheManger cachemanager.CacheManager, rpc LeaderboardScoringRPC, lbClient *leaderboardscoring.Client) Service {
	return Service{
		repository:            repo,
		validator:             validator,
		cacheManager:          cacheManger,
		leaderboardScoringRPC: rpc,
		lbScoringClient:       lbClient,
	}
}

func (s *Service) GetDailyContributorScores(ctx context.Context) error {
	log := logger.L()
	log.Info("Starting daily contributor scores calculation")

	if s.lbScoringClient == nil {
		return fmt.Errorf("leaderboardscoring client is not initialized")
	}

	var allDailyScores []DailyContributorScore
	pageSize := int32(1000) // TODO- Adjust based on what the service can handle
	offset := int32(0)

	for {
		// Get daily leaderboard from LeaderboardScoring service
		getLeaderboardReq := &lbscoring.GetLeaderboardRequest{
			Timeframe: "daily", // TODO - set proper timestamp
			PageSize:  pageSize,
			Offset:    offset,
		}

		leaderboardRes, err := s.lbScoringClient.GetLeaderboard(ctx, getLeaderboardReq)
		if err != nil {
			return fmt.Errorf("failed to get leaderboard data at offset %d: %w", offset, err)
		}

		log.Info("Retrieved leaderboard data",
			slog.Int("row_count", len(leaderboardRes.LeaderboardRows)),
			slog.String("timeframe", string(leaderboardRes.Timeframe)),
		)

		var dailyScores []DailyContributorScore

		for _, row := range leaderboardRes.LeaderboardRows {
			contributorID, err := s.mapUserIDToContributorID(ctx, row.UserID)
			if err != nil {
				log.Warn("Failed to map user ID to contributor ID",
					slog.String("user_id", row.UserID),
					slog.String("error", err.Error()),
				)

				continue
			}

			dailyScore := DailyContributorScore{
				ContributorID: contributorID,
				UserID:        row.UserID,
				Score:         float64(row.Score), // TODO - define is score data type float or int
				Rank:          row.Rank,           // TODO
				Timeframe:     string(leaderboardRes.Timeframe),
				//ProjectID:     1, // TODO add project_id to row response
			}
			dailyScores = append(dailyScores, dailyScore)
		}

		allDailyScores = append(allDailyScores, dailyScores...)
		if len(leaderboardRes.LeaderboardRows) < int(pageSize) {
			break
		}

		offset += pageSize

		time.Sleep(100 * time.Millisecond)
	}

	if len(allDailyScores) == 0 {
		log.Info("No leaderboard data found for daily calculation")
		return nil
	}

	if err := s.repository.StoreDailyContributorScores(ctx, allDailyScores); err != nil {
		return fmt.Errorf("failed to store daily contributor scores: %w", err)
	}

	go s.processDailyScoreCalculations(ctx, nil) // , allDailyScores

	// TODO - cache current day scores in anther job
	// TODO - cache key value pattern
	//if err := s.updateCacheAfterDailyCalculation(ctx, allDailyScores); err != nil {
	//	log.Warn("Failed to update cache after daily calculation", slog.String("error", err.Error()))
	//}

	log.Info("Successfully calculated and stored daily contributor scores",
		slog.Int("processed_count", len(allDailyScores)),
	)

	return nil
}

func (s *Service) processDailyScoreCalculations(ctx context.Context, dailyScores []DailyContributorScore) error {
	log := logger.L()
	log.Info("Starting background processing of daily score calculations")

	var pendingScores []DailyContributorScore
	if dailyScores == nil {
		// Get pending daily scores (status = 0)
		var err error
		pendingScores, err = s.repository.GetPendingDailyScores(ctx)
		if err != nil {
			return fmt.Errorf("failed to get pending daily scores: %w", err)
		}
	} else {
		pendingScores = dailyScores
	}

	if len(pendingScores) == 0 {
		log.Info("No pending daily scores to process")
		return nil
	}

	log.Info("Processing pending daily scores", slog.Int("count", len(pendingScores)))

	userProjectScores, err := s.calculateUserProjectScores(ctx, pendingScores)
	if err != nil {
		return fmt.Errorf("failed to calculate project scores: %w", err)
	}

	if err := s.repository.UpdateUserProjectScores(ctx, userProjectScores); err != nil {
		return fmt.Errorf("failed to update project scores: %w", err)
	}

	if err := s.repository.UpdateGlobalScores(ctx, userProjectScores); err != nil {
		return fmt.Errorf("failed to update global scores: %w", err)
	}

	// Mark daily scores as processed
	scoreIDs := make([]types.ID, len(pendingScores))
	for i, score := range pendingScores {
		if score.ID == 0 {
			continue
		}
		scoreIDs[i] = score.ID
	}

	if len(scoreIDs) == 0 {
		return nil
	}

	if err := s.repository.MarkDailyScoresAsProcessed(ctx, scoreIDs); err != nil {
		return fmt.Errorf("failed to mark daily scores as processed: %w", err)
	}

	if err := s.updateCacheAfterDailyCalculation(ctx, pendingScores); err != nil {
		log.Warn("Failed to update cache after daily calculation", slog.String("error", err.Error()))
	}

	log.Info("Successfully processed daily score calculations",
		slog.Int("projects_updated", len(userProjectScores)),
	)

	return nil
}

func (s *Service) calculateUserProjectScores(ctx context.Context, dailyScores []DailyContributorScore) ([]UserProjectScore, error) {
	// TODO - Implement project mapping logic
	sortedUserProjects := s.sortDailyScores(dailyScores)

	// to store sums: userID -> projectID -> total score
	userProjectSums := make(map[types.ID]map[types.ID]float64)

	for _, score := range sortedUserProjects {
		if userProjectSums[score.ContributorID] == nil {
			userProjectSums[score.ContributorID] = make(map[types.ID]float64)
		}
		userProjectSums[score.ContributorID][score.ProjectID] += score.Score
	}

	var userProjectScores []UserProjectScore
	for contributorID, projectMap := range userProjectSums {
		for projectID, totalScore := range projectMap {
			userProjectScores = append(userProjectScores, UserProjectScore{
				ContributorID: contributorID, // TODO-  Convert string to types.ID if needed
				ProjectID:     projectID,
				Score:         totalScore,
				Timeframe:     "daily",
				TimeValue:     fmt.Sprintf("%d/%d/%d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
			})
		}
	}

	return userProjectScores, nil
}

func (s *Service) sortDailyScores(dailyScores []DailyContributorScore) []DailyContributorScore {
	sorted := make([]DailyContributorScore, len(dailyScores))
	copy(sorted, dailyScores)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].ContributorID != sorted[j].ContributorID {
			return sorted[i].ContributorID < sorted[j].ContributorID
		}
		return sorted[i].ProjectID < sorted[j].ProjectID
	})

	return sorted
}

// TODO - Implement UserID To ContributorID
func (s *Service) mapUserIDToContributorID(ctx context.Context, userID string) (types.ID, error) {
	return types.ID(1), nil
}

func (s *Service) updateCacheAfterDailyCalculation(ctx context.Context, scores []DailyContributorScore) error {
	log := logger.L()
	cacheKey := "global_leaderboard:daily"
	if err := s.cacheManager.Delete(ctx, cacheKey); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	cacheSetErr := s.cacheManager.Set(ctx, cacheKey, scores, 24*time.Hour)
	if cacheSetErr != nil {
		return fmt.Errorf("failed to set cache: %w", cacheSetErr)
	}

	log.Info("Cache set successfully", slog.String("cache_key", cacheKey), slog.String("timestamp", time.Now().String()))
	return nil
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

/*func (s *Service) GetContributorTotalStats(ctx context.Context, contributorID types.ID) (ContributorTotalStats, error) {
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
}*/

func convertProtoProjectsScore(protoScores map[uint64]float64) map[types.ID]float64 {
	projectsScore := make(map[types.ID]float64)
	for projectID, score := range protoScores {
		projectsScore[types.ID(projectID)] = score
	}
	return projectsScore
}

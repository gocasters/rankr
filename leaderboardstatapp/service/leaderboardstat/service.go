package leaderboardstat

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/leaderboardscoring"
	lbscoring "github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/gommon/log"
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
	UpdateUserScores(ctx context.Context, userScores []UserScore) error
	UpdateProjectScores(ctx context.Context, projectScores []ProjectScore) error
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

func (s *Service) CalculateDailyContributorScores(ctx context.Context) error {
	log := logger.L()
	log.Info("Starting daily contributor scores calculation")

	if s.lbScoringClient == nil {
		return fmt.Errorf("leaderboardscoring client is not initialized")
	}

	// Get daily leaderboard from leaderboardscoring service
	getLeaderboardReq := &lbscoring.GetLeaderboardRequest{
		Timeframe: "", //leaderboardscoring.Daily, // TODO - set proper timestamp
		PageSize:  1000,
		Offset:    0,
	}

	leaderboardRes, err := s.lbScoringClient.GetLeaderboard(ctx, getLeaderboardReq)
	if err != nil {
		return fmt.Errorf("failed to get leaderboard data: %w", err)
	}

	log.Info("Retrieved leaderboard data",
		slog.Int("row_count", len(leaderboardRes.LeaderboardRows)),
		slog.String("timeframe", string(leaderboardRes.Timeframe)),
	)

	// Process and store daily scores
	var dailyScores []DailyContributorScore
	calculatedAt := time.Now()

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
			DailyScore:    float64(row.Score), // TODO - define is score data type float or int
			Rank:          row.Rank,
			Timeframe:     string(leaderboardRes.Timeframe),
			CalculatedAt:  calculatedAt,
		}
		dailyScores = append(dailyScores, dailyScore)
	}

	// Store the daily scores
	if err := s.repository.StoreDailyContributorScores(ctx, dailyScores); err != nil {
		return fmt.Errorf("failed to store daily contributor scores: %w", err)
	}

	// TODO- do calculations
	go s.processDailyScoreCalculations(context.Background(), dailyScores)

	// TODO - cache if it is needed
	// TODO - cache key value pattern
	//if err := s.updateCacheAfterDailyCalculation(ctx, dailyScores); err != nil {
	//	log.Warn("Failed to update cache after daily calculation", slog.String("error", err.Error()))
	//}

	log.Info("Successfully calculated and stored daily contributor scores",
		slog.Int("processed_count", len(dailyScores)),
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

	// Calculate user total scores
	userScores := s.calculateUserTotalScores(pendingScores)

	// Update user scores
	if err := s.repository.UpdateUserScores(ctx, userScores); err != nil {
		return fmt.Errorf("failed to update user scores: %w", err)
	}

	// Calculate project scores
	projectScores, err := s.calculateProjectScores(ctx, pendingScores)
	if err != nil {
		return fmt.Errorf("failed to calculate project scores: %w", err)
	}

	// Update project scores
	if err := s.repository.UpdateProjectScores(ctx, projectScores); err != nil {
		return fmt.Errorf("failed to update project scores: %w", err)
	}

	// Mark daily scores as processed
	scoreIDs := make([]types.ID, len(pendingScores))
	for i, score := range pendingScores {
		scoreIDs[i] = score.ID
	}
	if err := s.repository.MarkDailyScoresAsProcessed(ctx, scoreIDs); err != nil {
		return fmt.Errorf("failed to mark daily scores as processed: %w", err)
	}

	// Update cache
	if err := s.updateCacheAfterDailyCalculation(ctx, pendingScores); err != nil {
		log.Warn("Failed to update cache after daily calculation", slog.String("error", err.Error()))
	}

	log.Info("Successfully processed daily score calculations",
		slog.Int("users_updated", len(userScores)),
		slog.Int("projects_updated", len(projectScores)),
	)

	return nil
}

func (s *Service) calculateUserTotalScores(dailyScores []DailyContributorScore) []UserScore {
	userScoreMap := make(map[types.ID]float64)

	for _, score := range dailyScores {
		userScoreMap[score.ContributorID] += score.DailyScore
	}

	userScores := make([]UserScore, 0, len(userScoreMap))
	for userID, totalScore := range userScoreMap {
		userScores = append(userScores, UserScore{
			UserID: userID,
			Score:  totalScore,
		})
	}

	return userScores
}

func (s *Service) calculateProjectScores(ctx context.Context, dailyScores []DailyContributorScore) ([]ProjectScore, error) {
	// TODO - Implement project mapping logic

	//userProjects, err := s.getUserProjects(ctx, score.UserID)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get user projects: %w", err)
	//}
	sortedUserProjects := s.sortDailyScores(dailyScores)

	// Create a map to store sums: userID -> projectID -> total score
	userProjectSums := make(map[string]map[types.ID]float64)

	// Calculate sums
	for _, score := range sortedUserProjects {
		// Initialize user map if not exists
		if userProjectSums[score.UserID] == nil {
			userProjectSums[score.UserID] = make(map[types.ID]float64)
		}
		// Add score to the project total
		userProjectSums[score.UserID][score.ProjectID] += score.DailyScore
	}

	// Convert the map to ProjectScore slice
	var projectScores []ProjectScore
	for userID, projectMap := range userProjectSums {
		for projectID, totalScore := range projectMap {
			projectScores = append(projectScores, ProjectScore{
				UserID:    userID, // TODO-  Convert string to types.ID if needed
				ProjectID: projectID,
				Score:     totalScore,
			})
		}
	}

	return projectScores, nil
}

func (s *Service) sortDailyScores(dailyScores []DailyContributorScore) []DailyContributorScore {
	sorted := make([]DailyContributorScore, len(dailyScores))
	copy(sorted, dailyScores)

	// Sort by UserID (primary) and ProjectID (secondary)
	sort.Slice(sorted, func(i, j int) bool {
		// First compare by UserID
		if sorted[i].UserID != sorted[j].UserID {
			return sorted[i].UserID < sorted[j].UserID
		}
		// If UserID is the same, compare by ProjectID
		return sorted[i].ProjectID < sorted[j].ProjectID
	})

	return sorted
}

//func (s *Service) getUserProjects(ctx context.Context, userID string) ([]types.ID, error) {
//	// TODO: Implement actual logic to get user's projects
//	return []types.ID{types.ID(1), types.ID(2)}, nil
//}

func (s *Service) mapUserIDToContributorID(ctx context.Context, userID string) (types.ID, error) {
	return types.ID(1), nil
}

func (s *Service) updateCacheAfterDailyCalculation(ctx context.Context, scores []DailyContributorScore) error {
	cacheKey := "global_leaderboard:daily"
	if err := s.cacheManager.Delete(ctx, cacheKey); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	cacheSetErr := s.cacheManager.Set(ctx, cacheKey, scores, 24*time.Hour)
	if cacheSetErr != nil {
		return fmt.Errorf("failed to set cache: %w", cacheSetErr)
	}

	log.Info("Cache set successfully", slog.String(time.Now().String(), cacheKey))
	return nil
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

package leaderboardscoring

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/timettl"
	"log/slog"
)

var Timeframes = []string{"yearly", "monthly", "weekly"}

type RedisRepository interface {
	UpdateScores(ctx context.Context, keys []string, score int, userID string) error
}

type DatabaseRepository interface {
	PersistContribution(ctx context.Context, event *ContributionEvent) error
}

type Repository interface {
	RedisRepository
	DatabaseRepository
}

type Service struct {
	repo      Repository
	validator Validator
	logger    *slog.Logger
}

func NewService(repo Repository, validator Validator, logger *slog.Logger) Service {
	return Service{
		repo:      repo,
		validator: validator,
		logger:    logger,
	}
}

func (s Service) ProcessScoreEvent(ctx context.Context, req EventRequest) error {
	if err := s.validator.ValidateContributionEvent(req); err != nil {
		// TODO - use Error pattern
		return err
	}

	contributionEvent := ContributionEvent{
		ID:              req.ID,
		Type:            ContributionType(""),
		EventName:       req.EventName,
		RepositoryID:    req.RepositoryID,
		RepositoryName:  req.RepositoryName,
		ContributorID:   req.ContributorID,
		ScoreValue:      0,
		SourceReference: req.SourceReference,
		Timestamp:       req.Timestamp.UTC(),
	}

	var keys = s.keys(contributionEvent.RepositoryName)

	if err := s.repo.UpdateScores(ctx, keys, contributionEvent.ScoreValue, contributionEvent.ContributorID); err != nil {
		s.logger.Error(ErrFailedToUpdateScores.Error(), slog.String("error", err.Error()))
		return err
	}

	if err := s.repo.PersistContribution(ctx, &contributionEvent); err != nil {
		s.logger.Error(ErrFailedToPersistContribution.Error(), slog.String("error", err.Error()))
		return err
	}

	s.logger.Debug(MsgSuccessfullyProcessedEvent, slog.String("event_id", contributionEvent.ID))

	return nil
}

// GetLeaderboard TODO - get Leaderboard data
func (s Service) GetLeaderboard(ctx context.Context, req GetLeaderboardRequest) (GetLeaderboardResponse, error) {
	return GetLeaderboardResponse{}, nil
}

// CreateLeaderboardSnapshot TODO - reads the current state of the all-time leaderboards from Redis
// and persists them to the database. This should be run periodically by a scheduler.
func (s Service) CreateLeaderboardSnapshot(ctx context.Context) error {
	return nil
}

// RestoreLeaderboardFromSnapshot TODO - rebuilds the Redis leaderboards from the latest
// snapshot stored in the database. This is typically called on service startup if Redis is empty.
func (s Service) RestoreLeaderboardFromSnapshot(ctx context.Context) error {
	return nil
}

// All Keys

// Global Leaderboards
// leaderboard:global:all_time	members(user_ids)
// leaderboard:global:yearly:{year}
// leaderboard:global:monthly:{year}-{month}
// leaderboard:global:weekly:{year}-W{week_number}

// Per-Project Leaderboards
// leaderboard:{project_id}:all_time
// leaderboard:{project_id}:yearly:{year}
// leaderboard:{project_id}:monthly:{year}-{month}
// leaderboard:{project_id}:weekly:{year}-W{week_number}
func (s Service) keys(projectID string) []string {
	globalKeys := make([]string, 0, 4)
	perProjectKeys := make([]string, 0, 4)

	globalKeys = append(globalKeys, getGlobalLeaderboardKey("all_time", ""))
	for _, tf := range Timeframes {
		var period string

		switch tf {
		case "yearly":
			period = timettl.GetYear()
		case "monthly":
			period = timettl.GetMonth()
		case "weekly":
			period = timettl.GetWeek()
		}

		key := getGlobalLeaderboardKey(tf, period)
		globalKeys = append(globalKeys, key)
	}

	perProjectKeys = append(perProjectKeys, getPerProjectLeaderboardKey(projectID, "all_time", ""))
	for _, tf := range Timeframes {
		var period string

		switch tf {
		case "yearly":
			period = timettl.GetYear()
		case "monthly":
			period = timettl.GetMonth()
		case "weekly":
			period = timettl.GetWeek()
		}

		key := getPerProjectLeaderboardKey(projectID, tf, period)
		perProjectKeys = append(perProjectKeys, key)
	}

	keys := append(globalKeys, perProjectKeys...)
	return keys
}

func getGlobalLeaderboardKey(timeframe string, period string) string {
	if timeframe == "all_time" {
		return fmt.Sprintf("leaderboard:global:%s", timeframe)
	}

	return fmt.Sprintf("leaderboard:global:%s:%s", timeframe, period)
}

func getPerProjectLeaderboardKey(project string, timeframe string, period string) string {
	if timeframe == "all_time" {
		return fmt.Sprintf("leaderboard:%s:%s", project, timeframe)
	}

	return fmt.Sprintf("leaderboard:%s:%s:%s", project, timeframe, period)
}

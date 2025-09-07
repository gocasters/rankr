package leaderboardscoring

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/timettl"
	"log/slog"
	"strconv"
)

// ScoreRepository handles the hot path: real-time score updates in Redis.
type ScoreRepository interface {
	UpsertScores(ctx context.Context, score *UpsertScore) error
}

// DatabaseRepository handles the cold path: persisting events to the database.
type DatabaseRepository interface {
}

type Repository interface {
	ScoreRepository
	DatabaseRepository
}

type Service struct {
	repo      Repository
	validator Validator
}

func NewService(repo Repository, validator Validator) Service {
	return Service{
		repo:      repo,
		validator: validator,
	}
}

// ProcessScoreEvent handles only the critical, real-time login.
func (s Service) ProcessScoreEvent(ctx context.Context, req *EventRequest) error {
	logger := logger.L()

	if err := s.validator.ValidateEvent(req); err != nil {
		return err
	}

	score := s.calculateScore(req)
	if score == nil {
		logger.Debug("unsupported event payload; skipping", slog.String("event_id", req.ID))
		return nil
	}

	if err := s.repo.UpsertScores(ctx, score); err != nil {
		logger.Error(ErrFailedToUpdateScores.Error(), slog.String("error", err.Error()))
		return err
	}

	logger.Debug(MsgSuccessfullyProcessedEvent, slog.String("event_id", req.ID))

	return nil
}

func (s Service) GetLeaderboard(ctx context.Context, req *GetLeaderboardRequest) (GetLeaderboardResponse, error) {
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

	globalKeys = append(globalKeys, getGlobalLeaderboardKey(AllTime, ""))
	for _, tf := range Timeframes {
		var period string

		switch tf {
		case Yearly:
			period = timettl.GetYear()
		case Monthly:
			period = timettl.GetMonth()
		case Weekly:
			period = timettl.GetWeek()
		}

		key := getGlobalLeaderboardKey(tf, period)
		globalKeys = append(globalKeys, key)
	}

	perProjectKeys = append(perProjectKeys, getPerProjectLeaderboardKey(projectID, AllTime, ""))
	for _, tf := range Timeframes {
		var period string

		switch tf {
		case Yearly:
			period = timettl.GetYear()
		case Monthly:
			period = timettl.GetMonth()
		case Weekly:
			period = timettl.GetWeek()
		}

		key := getPerProjectLeaderboardKey(projectID, tf, period)
		perProjectKeys = append(perProjectKeys, key)
	}

	keys := append(globalKeys, perProjectKeys...)
	return keys
}

func (s Service) calculateScore(req *EventRequest) *UpsertScore {
	var keys = s.keys(strconv.FormatUint(req.RepositoryID, 10))

	switch payload := req.Payload.(type) {
	case PullRequestOpenedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  1,
			UserID: strconv.FormatUint(payload.UserID, 10),
		}

	case PullRequestClosedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  2,
			UserID: strconv.FormatUint(payload.UserID, 10),
		}

	case PullRequestReviewPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  3,
			UserID: strconv.FormatUint(payload.ReviewerUserID, 10),
		}

	case IssueOpenedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  4,
			UserID: strconv.FormatUint(payload.UserID, 10),
		}

	case IssueCommentedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  5,
			UserID: strconv.FormatUint(payload.UserID, 10),
		}

	case IssueClosedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  6,
			UserID: strconv.FormatUint(payload.UserID, 10),
		}

	case PushPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  7,
			UserID: strconv.FormatUint(payload.UserID, 10),
		}

	default:
		return nil
	}
}

func getGlobalLeaderboardKey(timeframe Timeframe, period string) string {
	if timeframe == AllTime {
		return fmt.Sprintf("leaderboard:global:%s", timeframe.String())
	}

	return fmt.Sprintf("leaderboard:global:%s:%s", timeframe.String(), period)
}

func getPerProjectLeaderboardKey(project string, timeframe Timeframe, period string) string {
	if timeframe == AllTime {
		return fmt.Sprintf("leaderboard:%s:%s", project, timeframe.String())
	}

	return fmt.Sprintf("leaderboard:%s:%s:%s", project, timeframe.String(), period)
}

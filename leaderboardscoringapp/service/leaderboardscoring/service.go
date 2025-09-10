package leaderboardscoring

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/timettl"
	"log/slog"
	"strconv"
)

// ScoreRepository handles the hot path: real-time score updates in Redis.
type ScoreRepository interface {
	UpsertScores(ctx context.Context, score *UpsertScore) error
	GetLeaderboard(ctx context.Context, leaderboard *LeaderboardQuery) (LeaderboardQueryResult, error)
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
		return errors.Join(ErrInvalidEventRequest, err)
	}

	score := s.calculateScore(req)
	if score == nil {
		logger.Debug("unsupported event payload; skipping", slog.String("event_id", req.ID))
		return nil
	}

	if err := s.repo.UpsertScores(ctx, score); err != nil {
		logger.Error(ErrFailedToUpdateScores.Error(), slog.String("error", err.Error()))
		return errors.Join(ErrFailedToUpdateScores, err)
	}

	logger.Debug(MsgSuccessfullyProcessedEvent, slog.String("event_id", req.ID))

	return nil
}

func (s Service) GetLeaderboard(ctx context.Context, req *GetLeaderboardRequest) (GetLeaderboardResponse, error) {
	log := logger.L()
	log.Debug(
		"GetLeaderboard request received in service layer",
		slog.Any("request", req),
	)

	if err := s.validator.ValidateGetLeaderboard(req); err != nil {
		log.Warn("Invalid leaderboard request", slog.String("error", err.Error()))
		return GetLeaderboardResponse{}, errors.Join(ErrInvalidArguments, err) // fmt.Errorf("%w:%v", ErrInvalidArguments, err)
	}

	key := req.BuildKey()

	const maxPageSize = 1000 // TODO: consider making this configurable
	if req.Offset < 0 {
		return GetLeaderboardResponse{}, errors.Join(ErrInvalidArguments, fmt.Errorf("offset must be >= 0"))
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	stop := int64(req.Offset) + int64(pageSize) - 1

	lbQuery := &LeaderboardQuery{
		Key:   key,
		Start: int64(req.Offset),
		Stop:  stop,
	}

	leaderboardScoring, err := s.repo.GetLeaderboard(ctx, lbQuery)
	if err != nil {
		log.Error("Failed to get leaderboard from repository", slog.String("error", err.Error()))
		return GetLeaderboardResponse{}, err
	}

	if len(leaderboardScoring.LeaderboardRows) == 0 {
		log.Debug("No leaderboard data found for the given criteria", slog.String("key", key))
		return GetLeaderboardResponse{}, ErrLeaderboardNotFound
	}

	leaderboardRes := mapLeaderboardScoringToParam(leaderboardScoring)
	leaderboardRes.Timeframe = req.Timeframe
	leaderboardRes.ProjectID = req.ProjectID

	log.Debug("Successfully retrieved leaderboard data", slog.Int("row_count", len(leaderboardRes.LeaderboardRows)))
	return leaderboardRes, nil
}

func mapLeaderboardScoringToParam(scoring LeaderboardQueryResult) GetLeaderboardResponse {
	leaderboardRes := GetLeaderboardResponse{
		Timeframe:       0,
		ProjectID:       nil,
		LeaderboardRows: make([]LeaderboardRow, 0, len(scoring.LeaderboardRows)),
	}

	for _, r := range scoring.LeaderboardRows {
		row := LeaderboardRow{
			Rank:   r.Rank,
			UserID: r.UserID,
			Score:  r.Score,
		}

		leaderboardRes.LeaderboardRows = append(leaderboardRes.LeaderboardRows, row)
	}

	return leaderboardRes
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

package leaderboardscoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/timettl"
	"log/slog"
	"strconv"
	"time"
)

// EventPersistence = database layer
type EventPersistence interface {
	AddProcessedScoreEvents(ctx context.Context, events []ProcessedScoreEvent) error
	AddUserTotalScores(ctx context.Context, snapshots []UserTotalScore) error
}

// LeaderboardCache = redis layer
type LeaderboardCache interface {
	UpsertScores(ctx context.Context, score *UpsertScore) error
	GetLeaderboard(ctx context.Context, leaderboard *LeaderboardQuery) (LeaderboardQueryResult, error)
}

// Publisher interface for publishing processed events
type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

type Service struct {
	eventPersistence    EventPersistence
	leaderboard         LeaderboardCache
	publisher           Publisher
	processedEventTopic string
	validator           Validator
}

func NewService(
	persistence EventPersistence,
	leaderboard LeaderboardCache,
	publisher Publisher,
	processedEventTopic string,
	validator Validator,
) *Service {
	return &Service{
		eventPersistence:    persistence,
		leaderboard:         leaderboard,
		publisher:           publisher,
		processedEventTopic: processedEventTopic,
		validator:           validator,
	}
}

// ProcessScoreEvent handles only the critical, real-time login.
func (s *Service) ProcessScoreEvent(ctx context.Context, req *EventRequest) error {
	log := logger.L()

	if err := s.validator.ValidateEvent(req); err != nil {
		return errors.Join(ErrInvalidEventRequest, err)
	}

	score := calculateScore(req.EventName)
	if score == 0 {
		log.Debug("unsupported event payload; skipping", slog.String("event_id", req.ID))
		return nil
	}

	// Update Redis leaderboard (real-time)
	keys := s.keys(strconv.FormatUint(req.RepositoryID, 10))
	upsertScore := UpsertScore{
		Keys:   keys,
		Score:  score,
		UserID: req.UserID,
	}
	if err := s.leaderboard.UpsertScores(ctx, &upsertScore); err != nil {
		log.Error(ErrFailedToUpdateScores.Error(), slog.String("error", err.Error()))
		return errors.Join(ErrFailedToUpdateScores, err)
	}

	// Publish to NATS JetStream for batch persistence
	pse := ProcessedScoreEvent{
		UserID:    req.UserID,
		EventName: EventName(req.EventName),
		Score:     score,
		Timestamp: time.Now().UTC(),
	}

	dataMsg, mErr := json.Marshal(pse)
	if mErr != nil {
		log.Error("failed to marshal processed score event", slog.String("error", mErr.Error()))
		return fmt.Errorf("marshal event: %w", mErr)
	}

	if err := s.publisher.Publish(ctx, s.processedEventTopic, dataMsg); err != nil {
		log.Error("failed to publish processed score event", slog.String("error", err.Error()))
		return fmt.Errorf("publish event: %w", err)
	}

	log.Debug(MsgSuccessfullyProcessedEvent, slog.String("event_id", req.ID))
	return nil
}

func (s *Service) GetLeaderboard(ctx context.Context, req *GetLeaderboardRequest) (GetLeaderboardResponse, error) {
	log := logger.L()
	log.Debug(
		"GetLeaderboard request received in service layer",
		slog.Any("request", req),
	)

	if err := s.validator.ValidateGetLeaderboard(req); err != nil {
		log.Warn("Invalid leaderboard request", slog.String("error", err.Error()))
		return GetLeaderboardResponse{}, errors.Join(ErrInvalidArguments, err)
	}

	key := req.BuildKey()

	stop := int64(req.Offset) + int64(req.PageSize) - 1

	lbQuery := &LeaderboardQuery{
		Key:   key,
		Start: int64(req.Offset),
		Stop:  stop,
	}

	leaderboardScoring, err := s.leaderboard.GetLeaderboard(ctx, lbQuery)
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

// CreateLeaderboardSnapshot TODO - reads the current state of the all-time leaderboards from Redis
// and persists them to the database. This should be run periodically by a scheduler.
func (s *Service) CreateLeaderboardSnapshot(ctx context.Context) error {
	return nil
}

// RestoreLeaderboardFromSnapshot TODO - rebuilds the Redis leaderboards from the latest
// snapshot stored in the database. This is typically called on service startup if Redis is empty.
func (s *Service) RestoreLeaderboardFromSnapshot(ctx context.Context) error {
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
func (s *Service) keys(projectID string) []string {
	keyMap := make(map[string]struct{})
	addKey := func(key string) {
		if _, exists := keyMap[key]; !exists {
			keyMap[key] = struct{}{}
		}
	}

	// Global keys
	addKey(getGlobalLeaderboardKey(AllTime, ""))
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
		addKey(getGlobalLeaderboardKey(tf, period))
	}

	// Per-project keys
	addKey(getPerProjectLeaderboardKey(projectID, AllTime, ""))
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
		addKey(getPerProjectLeaderboardKey(projectID, tf, period))
	}

	// Convert map keys to slice
	keys := make([]string, 0, len(keyMap))
	for k := range keyMap {
		keys = append(keys, k)
	}

	return keys
}

func calculateScore(eventType string) int64 {
	//var keys = s.keys(strconv.FormatUint(req.RepositoryID, 10))
	switch eventType {
	case PullRequestOpened.String():
		return 1
	case PullRequestClosed.String():
		return 2
	case PullRequestReview.String():
		return 3
	case IssueOpened.String():
		return 4
	case IssueClosed.String():
		return 5
	case IssueComment.String():
		return 6
	case CommitPush.String():
		return 7
	default:
		return 0
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

func mapLeaderboardScoringToParam(scoring LeaderboardQueryResult) GetLeaderboardResponse {
	leaderboardRes := GetLeaderboardResponse{
		Timeframe:       TimeframeUnspecified.String(),
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

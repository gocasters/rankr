package leaderboardscoring

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/timettl"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"strconv"
)

// ScoreRepository handles the hot path: real-time score updates in Redis.
type ScoreRepository interface {
	UpsertScores(ctx context.Context, score *UpsertScore) error
}

// PersistenceQueueRepository handles the temporary buffering of events.
// TODO- This could be implemented with Redis Lists, Kafka, etc.
type PersistenceQueueRepository interface {
	Enqueue(ctx context.Context, payload []byte) error
	DequeueBatch(ctx context.Context, batchSize int) ([][]byte, error)
}

// DatabaseRepository handles the cold path: persisting events to the database.
type DatabaseRepository interface {
	PersistEventBatch(ctx context.Context, events []*Event) error
}

type Repository interface {
	ScoreRepository
	PersistenceQueueRepository
	DatabaseRepository
}

type Config struct {
	BatchSize int `koanf:"batch_size"`
}

type Service struct {
	repo      Repository
	validator Validator
	cfg       Config
}

func NewService(repo Repository, validator Validator, cfg Config) Service {
	return Service{
		repo:      repo,
		validator: validator,
		cfg:       cfg,
	}
}

// ProcessScoreEvent handles only the critical, real-time login.
func (s Service) ProcessScoreEvent(ctx context.Context, req *EventRequest) error {
	logger := logger.L()

	if err := s.validator.ValidateEvent(req); err != nil {
		return err
	}

	score := s.calculateScore(req)

	if err := s.repo.UpsertScores(ctx, score); err != nil {
		logger.Error(ErrFailedToUpdateScores.Error(), slog.String("error", err.Error()))
		return err
	}

	logger.Debug(MsgSuccessfullyProcessedEvent, slog.String("event_id", req.ID))

	return nil
}

// QueueEventForPersistence handles the non-critical, async part.
func (s Service) QueueEventForPersistence(ctx context.Context, eventPayload []byte) error {
	return s.repo.Enqueue(ctx, eventPayload)
}

// ProcessPersistenceQueue is the method called by the scheduler for persist async Event to Database.
func (s Service) ProcessPersistenceQueue(ctx context.Context) error {
	logger := logger.L()
	rawEvents, err := s.repo.DequeueBatch(ctx, s.cfg.BatchSize)
	if err != nil {
		logger.Error("failed to dequeue events from persistence queue", slog.String("error", err.Error()))
		return err
	}
	if len(rawEvents) == 0 {
		logger.Debug("no events in persistence queue to process.")
		return nil
	}

	var events = make([]*Event, 0, len(rawEvents))
	for _, payload := range rawEvents {
		var protoEvent eventpb.Event
		if err := proto.Unmarshal(payload, &protoEvent); err != nil {
			logger.Error("failed to unmarshal event from queue, skipping", slog.String("error", err.Error()))
			continue
		}

		// TODO - Map the protoEvent to internal Event struct.
		// event := MapProtoToDomain(&protoEvent)
		// events = append(events, event)
	}

	// TODO - map batchPayloadEvent to []Event,
	if err := s.repo.PersistEventBatch(ctx, events); err != nil {
		logger.Error("failed to persist batch of events to database", slog.String("error", err.Error()))
		// TODO - Implement a dead-letter queue or retry mechanism for the failed batch.
		return err
	}

	logger.Info("successfully persisted event batch to database", slog.Int("batch_size", len(events)))

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
	var keys = s.keys(strconv.Itoa(int(req.RepositoryID)))

	switch payload := req.Payload.(type) {
	case PullRequestOpenedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  1,
			UserID: strconv.Itoa(int(payload.UserID)),
		}

	case PullRequestClosedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  2,
			UserID: strconv.Itoa(int(payload.UserID)),
		}

	case PullRequestReviewPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  3,
			UserID: strconv.Itoa(int(payload.ReviewerUserID)),
		}

	case IssueOpenedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  4,
			UserID: strconv.Itoa(int(payload.UserID)),
		}

	case IssueCommentedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  5,
			UserID: strconv.Itoa(int(payload.UserID)),
		}

	case IssueClosedPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  6,
			UserID: strconv.Itoa(int(payload.UserID)),
		}

	case PushPayload:
		return &UpsertScore{
			Keys:   keys,
			Score:  7,
			UserID: strconv.Itoa(int(payload.UserID)),
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

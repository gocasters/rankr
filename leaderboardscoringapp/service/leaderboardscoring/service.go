package leaderboardscoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/timettl"
)

const (
	snapshotBatchSize     = 100  // Fetch 100 users at a time
	snapshotPersistBatch  = 1000 // Persist every 1000 rows
	snapshotRetryAttempts = 3
)

// EventPersistence = database layer
type EventPersistence interface {
	AddProcessedScoreEvents(ctx context.Context, events []ProcessedScoreEvent) error
	AddSnapshot(ctx context.Context, snapshots []SnapshotRow) error
}

// LeaderboardCache = redis layer
type LeaderboardCache interface {
	UpsertScores(ctx context.Context, score *UpsertScore, timeframe Timeframe) error
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

	// Update Redis leaderboard (real-time) for all timeframes
	for _, tf := range Timeframes {
		keys := s.generateKeys(strconv.FormatUint(req.RepositoryID, 10), tf)

		upsertScore := UpsertScore{
			Keys:   keys,
			Score:  score,
			UserID: req.UserID,
		}
		if err := s.leaderboard.UpsertScores(ctx, &upsertScore, tf); err != nil {
			log.Error(ErrFailedToUpdateScores.Error(), slog.String("error", err.Error()))
			return errors.Join(ErrFailedToUpdateScores, err)
		}
	}

	// Publish to NATS JetStream for batch persistence (once per event)
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
		return GetLeaderboardResponse{}, nil
	}

	leaderboardRes := mapLeaderboardScoringToParam(leaderboardScoring)
	leaderboardRes.Timeframe = req.Timeframe
	leaderboardRes.ProjectID = req.ProjectID

	log.Debug("Successfully retrieved leaderboard data", slog.Int("row_count", len(leaderboardRes.LeaderboardRows)))
	return leaderboardRes, nil
}

// LeaderboardSnapshot creates snapshots for specified project leaderboards
func (s *Service) LeaderboardSnapshot(ctx context.Context, projectIDs []string) error {
	log := logger.L()

	log.Info("starting leaderboard snapshot creation", slog.Int("project_count", len(projectIDs)))

	keys := getSnapshotKeys(projectIDs)
	if len(keys) == 0 {
		log.Warn("no snapshot keys to process")
		return nil
	}

	// Process keys
	results := s.processKeys(ctx, keys)

	var errs []error
	for err := range results {
		if err != nil {
			errs = append(errs, err)
			log.Error("snapshot failed", slog.String("error", err.Error()))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("snapshot had %d failures: %v", len(errs), errs)
	}

	return nil
}

// processKeys processes multiple keys concurrently
func (s *Service) processKeys(ctx context.Context, keys []string) <-chan error {
	results := make(chan error, len(keys))
	var wg sync.WaitGroup

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()

			_, err := s.createSnapshotForKey(ctx, k)
			results <- err
		}(key)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

// createSnapshotForKey creates snapshot for a single leaderboard key
func (s *Service) createSnapshotForKey(ctx context.Context, key string) (int, error) {
	log := logger.L()
	snapshotTime := time.Now().UTC()

	var tempSnapshots []SnapshotRow
	totalRows := 0

	offset := int64(0)
	batchSize := int64(snapshotBatchSize)

	for {
		select {
		case <-ctx.Done():
			return totalRows, ctx.Err()
		default:
		}

		// Fetch batch from Redis
		query := &LeaderboardQuery{
			Key:   key,
			Start: offset,
			Stop:  offset + batchSize - 1,
		}

		leaderboard, err := s.leaderboard.GetLeaderboard(ctx, query)
		if err != nil {
			return totalRows, fmt.Errorf("failed to fetch leaderboard: %w", err)
		}

		if len(leaderboard.LeaderboardRows) == 0 {
			break
		}

		// Convert to snapshot rows
		for _, row := range leaderboard.LeaderboardRows {
			snapshot := SnapshotRow{
				Rank:              row.Rank,
				UserID:            row.UserID,
				TotalScore:        row.Score,
				LeaderboardKey:    key,
				SnapshotTimestamp: snapshotTime,
			}
			tempSnapshots = append(tempSnapshots, snapshot)
		}

		totalRows += len(leaderboard.LeaderboardRows)

		if len(tempSnapshots) >= snapshotPersistBatch {
			if err := s.persistSnapshotBatch(ctx, tempSnapshots); err != nil {
				return totalRows, fmt.Errorf("failed to persist snapshot batch: %w", err)
			}
			log.Debug("persisted snapshot batch",
				slog.String("key", key),
				slog.Int("batch_size", len(tempSnapshots)))
			tempSnapshots = tempSnapshots[:0] // Reset slice
		}

		// Next batch
		offset += batchSize

		// Safety check: prevent infinite loop
		if offset > maxOffset {
			log.Warn("offset exceeded safety limit, stopping pagination", slog.String("key", key))
			break
		}
	}

	if len(tempSnapshots) > 0 {
		if err := s.persistSnapshotBatch(ctx, tempSnapshots); err != nil {
			return totalRows, fmt.Errorf("failed to persist final snapshot batch: %w", err)
		}
	}

	return totalRows, nil
}

// persistSnapshotBatch persists a batch with retry logic
func (s *Service) persistSnapshotBatch(ctx context.Context, snapshots []SnapshotRow) error {
	log := logger.L()

	var lastErr error
	for attempt := 1; attempt <= snapshotRetryAttempts; attempt++ {
		err := s.eventPersistence.AddSnapshot(ctx, snapshots)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Warn("failed persist snapshots",
			slog.String("error", lastErr.Error()))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * 2 * time.Second):
		}
	}

	return lastErr
}

// RestoreLeaderboardFromSnapshot rebuilds Redis leaderboards from the latest snapshot
func (s *Service) RestoreLeaderboardFromSnapshot(ctx context.Context, projectIDs []string) error {
	log := logger.L()

	keys := getSnapshotKeys(projectIDs)
	log.Info("starting leaderboard restore from snapshot", slog.Int("keys", len(keys)))

	// TODO: Implement restore logic
	// 1. Query latest snapshots for each key from database
	// 2. Group by leaderboard_key
	// 3. Bulk insert to Redis using pipeline
	// 4. Verify counts match

	return fmt.Errorf("not implemented")
}

// Helper: get snapshot keys for projects
func getSnapshotKeys(projectIDs []string) []string {
	keys := make([]string, 0, len(projectIDs)+1)

	// Global all-time leaderboard
	keys = append(keys, getGlobalLeaderboardKey(AllTime, ""))

	// Per-project all-time leaderboards
	for _, pID := range projectIDs {
		if pID != "" {
			keys = append(keys, getPerProjectLeaderboardKey(pID, AllTime, ""))
		}
	}

	return keys
}

// All Keys

// Global Leaderboards
// leaderboard:global:all_time	members(user_ids)
// leaderboard:global:yearly:{year}
// leaderboard:global:monthly:{year}-{month}
// leaderboard:global:weekly:{year}-W{week_number}
// leaderboard:global:daily:{year}-{week_number}-{day_number}

// Per-Project Leaderboards
// leaderboard:{project_id}:all_time
// leaderboard:{project_id}:yearly:{year}
// leaderboard:{project_id}:monthly:{year}-{month}
// leaderboard:{project_id}:weekly:{year}-W{week_number}
// leaderboard:{project_id}:daily:{year}-{week_number}-{day_number}
func (s *Service) generateKeys(projectID string, timeframe Timeframe) []string {
	keyMap := make(map[string]struct{})
	addKey := func(key string) {
		if _, exists := keyMap[key]; !exists {
			keyMap[key] = struct{}{}
		}
	}

	var period string
	var tf Timeframe
	switch timeframe {
	case AllTime:
		addKey(getGlobalLeaderboardKey(AllTime, ""))
		addKey(getPerProjectLeaderboardKey(projectID, AllTime, ""))
	case Yearly:
		period = timettl.GetYear()
		tf = Yearly
	case Monthly:
		period = timettl.GetMonth()
		tf = Monthly
	case Weekly:
		period = timettl.GetWeek()
		tf = Weekly
	case Daily:
		period = timettl.GetDay()
		tf = Daily
	}

	if timeframe != AllTime {
		addKey(getGlobalLeaderboardKey(tf, period))
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

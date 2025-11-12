package postgrerepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
	"log/slog"
	"math/rand"
	"time"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RetryConfig struct {
	MaxRetries int           `koanf:"max_retries"`
	RetryDelay time.Duration `koanf:"retry_delay"`
}

// PostgreSQLRepository handles persistence of processed events and snapshots
type PostgreSQLRepository struct {
	postgreSQL  *database.Database
	retryConfig RetryConfig
}

func NewPostgreSQLRepository(db *database.Database, config RetryConfig) leaderboardscoring.EventPersistence {
	return &PostgreSQLRepository{
		postgreSQL:  db,
		retryConfig: config,
	}
}

func (db PostgreSQLRepository) AddProcessedScoreEvents(ctx context.Context, events []leaderboardscoring.ProcessedScoreEvent) error {
	if len(events) == 0 {
		return nil
	}

	return db.retryOperation(
		ctx,
		func() error {
			return db.insertBatchProcessedScoreEvent(ctx, events)
		},
	)
}

func (db PostgreSQLRepository) AddSnapshot(ctx context.Context, snapshots []leaderboardscoring.SnapshotRow) error {
	if len(snapshots) == 0 {
		return nil
	}

	return db.retryOperation(ctx, func() error {
		return db.insertBatchSnapshot(ctx, snapshots)
	})
}

// retryOperation - Generic retry logic
func (db PostgreSQLRepository) retryOperation(ctx context.Context, operation func() error) error {
	var lastErr error
	var defaultDelay = 100 * time.Millisecond

	// Ensure at least one attempt
	maxRetries := db.retryConfig.MaxRetries
	if maxRetries < 1 {
		maxRetries = 1
	}

	base := db.retryConfig.RetryDelay
	if base <= 0 {
		base = defaultDelay
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Exponential backoff: base, 2×base, 4×base, ...
			exp := base << (attempt - 2)

			// Jitter (random between 0 and exp
			jitter := time.Duration(rand.Int63n(int64(exp)))

			delay := exp + jitter

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		if !isRetryableError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (db PostgreSQLRepository) insertBatchProcessedScoreEvent(ctx context.Context, events []leaderboardscoring.ProcessedScoreEvent) error {
	tx, err := db.postgreSQL.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	columns := []string{"user_id", "event_type", "event_timestamp", "score_delta"}

	rows := make([][]interface{}, len(events))
	for i, event := range events {
		rows[i] = []interface{}{
			event.UserID,
			event.EventName.String(),
			event.Timestamp,
			event.Score,
		}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"processed_score_events"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy from: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// insertBatchSnapshot inserts multiple user score snapshots using a three-step process
// to ensure idempotency and optimal performance.
//
// Returns:
//   - nil on success (including when all snapshots were duplicates)
//   - error if database operation fails (connection, syntax, constraint violations other than duplicates)
func (db PostgreSQLRepository) insertBatchSnapshot(ctx context.Context, snapshots []leaderboardscoring.SnapshotRow) error {
	if len(snapshots) == 0 {
		return nil
	}

	tx, err := db.postgreSQL.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Step 1: Create temp table
	_, err = tx.Exec(ctx, `
        CREATE TEMP TABLE temp_snapshot (
            rank BIGINT NOT NULL,
            user_id VARCHAR(100) NOT NULL,
            total_score BIGINT NOT NULL,
			leaderboard_key VARCHAR(250) NOT NULL,
            snapshot_timestamp TIMESTAMP NOT NULL
        ) ON COMMIT DROP
    `)
	if err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	// Step 2: Bulk insert to temp table using CopyFrom
	columns := []string{"rank", "user_id", "total_score", "leaderboard_key", "snapshot_timestamp"}
	rows := make([][]interface{}, len(snapshots))
	for i, ss := range snapshots {
		rows[i] = []interface{}{
			ss.Rank,
			ss.UserID,
			ss.TotalScore,
			ss.LeaderboardKey,
			ss.SnapshotTimestamp,
		}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_snapshot"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy to temp table: %w", err)
	}

	// Step 3: Insert only new snapshots (ignore duplicates)
	_, err = tx.Exec(ctx, `
        INSERT INTO snapshot (rank, user_id, total_score, leaderboard_key, snapshot_timestamp)
        SELECT rank, user_id, total_score, leaderboard_key, snapshot_timestamp
        FROM temp_snapshot
        ON CONFLICT (user_id, leaderboard_key, snapshot_timestamp) DO NOTHING
    `)
	if err != nil {
		return fmt.Errorf("insert snapshots: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case statuscode.ErrCodeUniqueViolation:
			return false
		case statuscode.ErrCodeForeignKeyViolation:
			return false

		case statuscode.ErrCodeSerializationFailure,
			statuscode.ErrCodeDeadlockDetected:
			return true

		case statuscode.ErrCodeConnectionException,
			statuscode.ErrCodeConnectionNotExist,
			statuscode.ErrCodeConnectionFailure:
			logger.L().Info("postgres connection error; will retry",
				slog.String("code", pgErr.Code),
				slog.String("severity", pgErr.Severity),
				slog.String("message", pgErr.Message),
			)
			return true

		case statuscode.ErrTooManyConnections:
			return true
		}
	}

	return false
}

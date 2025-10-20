package postgrerepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
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

func (db PostgreSQLRepository) AddSnapshotTotalScores(ctx context.Context, snapshots []leaderboardscoring.UserTotalScore) error {
	if len(snapshots) == 0 {
		return nil
	}

	return db.retryOperation(ctx, func() error {
		return db.insertBatchSnapshotTotalScores(ctx, snapshots)
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

// insertBatchSnapshotTotalScores inserts multiple user score snapshots using a three-step process
// to ensure idempotency and optimal performance.
//
// This method is designed for hourly snapshot operations where duplicate snapshots
// (same user_id + snapshot_timestamp) should be silently ignored rather than causing errors.
//
// Example usage:
//
//	snapshots := []UserTotalScore{
//	    {UserID: "user123", TotalScore: 1500, SnapshotTimestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
//	    {UserID: "user456", TotalScore: 2300, SnapshotTimestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
//	}
//	err := repo.insertBatchUserTotalScores(ctx, snapshots)
//
// Process:
//  1. Creates a temporary table to stage the incoming snapshots
//  2. Uses PostgreSQL's COPY protocol for fast bulk insertion into temp table
//  3. Inserts from temp table to final table with ON CONFLICT DO NOTHING
//     - This ensures duplicates are ignored without causing transaction rollback
//     - Preserves historical snapshots as separate rows (no updates)
//
// Idempotency guarantee:
//   - First call:  INSERT snapshot (user1, score=100, timestamp=12:00) ✓ Success
//   - Second call: INSERT snapshot (user1, score=100, timestamp=12:00) ✓ Ignored (DO NOTHING)
//   - Third call:  INSERT snapshot (user1, score=150, timestamp=13:00) ✓ Success (new timestamp)
//
// Performance characteristics:
//   - Leverages PostgreSQL's COPY protocol (10-100x faster than individual INSERTs)
//   - Temp table approach avoids multiple round-trips to database
//   - Suitable for batches of thousands of snapshots
//
// Thread safety:
//   - Safe for concurrent calls due to transaction isolation
//   - Each transaction gets its own temporary table (ON COMMIT DROP)
//
// Returns:
//   - nil on success (including when all snapshots were duplicates)
//   - error if database operation fails (connection, syntax, constraint violations other than duplicates)
func (db PostgreSQLRepository) insertBatchSnapshotTotalScores(ctx context.Context, snapshots []leaderboardscoring.UserTotalScore) error {
	if len(snapshots) == 0 {
		return nil
	}

	tx, err := db.postgreSQL.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Step 1: Create temp table
	// The temp table is automatically dropped at transaction end (ON COMMIT DROP)
	// This ensures no leftover data and allows concurrent executions
	_, err = tx.Exec(ctx, `
        CREATE TEMP TABLE temp_user_snapshots (
            user_id VARCHAR(100) NOT NULL,
            total_score BIGINT NOT NULL,
            snapshot_timestamp TIMESTAMP NOT NULL
        ) ON COMMIT DROP
    `)
	if err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	// Step 2: Bulk insert to temp table using CopyFrom
	// CopyFrom uses PostgreSQL's COPY protocol which is significantly faster
	// than executing multiple INSERT statements
	columns := []string{"user_id", "total_score", "snapshot_timestamp"}
	rows := make([][]interface{}, len(snapshots))
	for i, ss := range snapshots {
		rows[i] = []interface{}{
			ss.UserID,
			ss.TotalScore,
			ss.SnapshotTimestamp,
		}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_user_snapshots"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy to temp table: %w", err)
	}

	// Step 3: Insert only new snapshots (ignore duplicates)
	// ON CONFLICT DO NOTHING ensures that if a snapshot with the same
	// (user_id, snapshot_timestamp) already exists, it will be silently skipped
	// This maintains idempotency - retrying the same operation won't cause errors
	_, err = tx.Exec(ctx, `
        INSERT INTO user_total_scores (user_id, total_score, snapshot_timestamp)
        SELECT user_id, total_score, snapshot_timestamp
        FROM temp_user_snapshots
        ON CONFLICT (user_id, snapshot_timestamp) DO NOTHING
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
			logger.L().Info("------------------> ", time.Now(), ":", pgErr.Code)
			return true
		case statuscode.ErrTooManyConnections:
			time.Sleep(5 * time.Second)
			return true
		}
	}

	return true
}

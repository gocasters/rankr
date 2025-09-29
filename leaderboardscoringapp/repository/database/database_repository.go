package postgrerepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL Error Codes
const (
	ErrCodeUniqueViolation      = "23505" // Duplicate key
	ErrCodeForeignKeyViolation  = "23503" // Invalid FK
	ErrCodeSerializationFailure = "40001" // Transaction conflict
	ErrCodeDeadlockDetected     = "40P01" // Deadlock
	ErrCodeConnectionException  = "08000" // Connection problem
	ErrCodeConnectionNotExist   = "08003" // Connection closed
	ErrCodeConnectionFailure    = "08006" // Connection failed
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

	return db.retryOperation(ctx, func() error {
		return db.insertBatchProcessedScoreEvent(ctx, events)
	})
}

func (db PostgreSQLRepository) AddUserTotalScores(ctx context.Context, snapshots []leaderboardscoring.UserTotalScore) error {
	if len(snapshots) == 0 {
		return nil
	}

	return db.retryOperation(ctx, func() error {
		return db.insertBatchUserTotalScores(ctx, snapshots)
	})
}

// retryOperation - Generic retry logic
func (db PostgreSQLRepository) retryOperation(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < db.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 300ms
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(db.retryConfig.RetryDelay * time.Duration(attempt)):
			}
		}

		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", db.retryConfig.MaxRetries, lastErr)
}

func (db PostgreSQLRepository) insertBatchProcessedScoreEvent(ctx context.Context, events []leaderboardscoring.ProcessedScoreEvent) error {
	tx, err := db.postgreSQL.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	columns := []string{"user_id", "event_type", "event_timestamp", "score_delta"}

	rows := make([][]interface{}, len(events))
	for i, event := range events {
		rows[i] = []interface{}{
			event.UserID,
			string(event.EventName),
			event.Timestamp,
			int(event.Score),
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

func (db PostgreSQLRepository) insertBatchUserTotalScores(ctx context.Context, snapshots []leaderboardscoring.UserTotalScore) error {
	tx, err := db.postgreSQL.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	columns := []string{"user_id", "total_score", "snapshot_timestamp"}

	rows := make([][]interface{}, len(snapshots))
	for i, ss := range snapshots {
		rows[i] = []interface{}{
			ss.UserID,
			int(ss.TotalScore),
			ss.SnapshotTimestamp,
		}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"user_total_scores"},
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

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors - don't retry
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	// PostgreSQL specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		// Constraint violations - data issues, don't retry
		case ErrCodeUniqueViolation:
			// Duplicate key - retrying won't help
			return false
		case ErrCodeForeignKeyViolation:
			// Invalid FK - data problem, don't retry
			return false

		// Transaction conflicts - temporary, should retry
		case ErrCodeSerializationFailure:
			// Concurrent transactions conflict - retry can succeed
			return true
		case ErrCodeDeadlockDetected:
			// Deadlock - one TX was killed, retry can succeed
			return true

		// Connection errors - temporary, should retry
		case ErrCodeConnectionException, ErrCodeConnectionNotExist, ErrCodeConnectionFailure:
			// Network/connection issues - retry can succeed
			return true
		}
	}

	// Default: retry for unknown errors
	// Conservative approach - let retry logic handle it
	return true
}

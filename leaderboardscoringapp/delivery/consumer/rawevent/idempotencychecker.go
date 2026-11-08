package rawevent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	ProcessedKeyTTL time.Duration `koanf:"processed_key_ttl"` // 24 * time.Hour
	LockKeyTTL      time.Duration `koanf:"lock_key_ttl"`      // 5 * time.Minute

	PrefixProcessedKey string `koanf:"prefix_processed_key"` // processed_events
	PrefixLockKey      string `koanf:"prefix_lock_key"`      // lock
}

type keyStore interface {
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	DeleteIfEqual(ctx context.Context, key, value string) error
}

type IdempotencyChecker struct {
	store  keyStore
	config Config
}

func NewIdempotencyChecker(client *redis.Client, config Config) *IdempotencyChecker {
	return &IdempotencyChecker{
		store:  redisKeyStore{client: client},
		config: config,
	}
}

var (
	ErrEventAlreadyProcessed = errors.New("event already processed")
	ErrEventLocked           = errors.New("event is currently locked by another processor")
)

const processedMarker = "1"

// CheckEvent It returns specific errors if the event is a duplicate or is locked.
func (ic *IdempotencyChecker) CheckEvent(ctx context.Context, eventID string, processEventFunc func() error) error {
	if eventID == "" {
		return fmt.Errorf("invalid eventID: empty")
	}

	processedKey := ic.processedKey(eventID)
	lockKey := ic.lockKey(eventID)

	// 1. Check if the event has already been successfully processed.
	exists, err := ic.store.Exists(ctx, processedKey)
	if err != nil {
		return fmt.Errorf("failed to check for processed event: %w", err)
	}
	if exists {
		return ErrEventAlreadyProcessed
	}

	// 2. Try to acquire a temporary lock.
	token := uuid.NewString()
	wasSet, sErr := ic.store.SetNX(ctx, lockKey, token, ic.config.LockKeyTTL)
	if sErr != nil {
		return fmt.Errorf("failed to acquire lock: %w", sErr)
	}
	if !wasSet {
		return ErrEventLocked
	}
	defer ic.releaseLock(ctx, lockKey, token)

	// 3. Execute the core business logic.
	if pErr := processEventFunc(); pErr != nil {
		return pErr
	}

	// 4. If successful, mark the event as permanently processed.
	if sErr := ic.store.Set(ctx, processedKey, processedMarker, ic.config.ProcessedKeyTTL); sErr != nil {
		// This is a critical failure. The event was processed, but we couldn't mark it as such.
		logger.L().Error(
			"CRITICAL: Failed to mark event as processed after successful execution",
			slog.String("event_id", eventID),
			slog.String("error", sErr.Error()),
		)

		return fmt.Errorf("critical: failed to mark event as processed: %w", sErr)
	}

	return nil
}

func (ic *IdempotencyChecker) releaseLock(ctx context.Context, lockKey, token string) {
	if err := ic.store.DeleteIfEqual(ctx, lockKey, token); err != nil {
		logger.L().Warn("failed to release lock (will expire by TTL)", slog.String("key", lockKey), slog.String("error", err.Error()))
	}
}

func (ic *IdempotencyChecker) processedKey(eventID string) string {
	return fmt.Sprintf("%s:%s", ic.config.PrefixProcessedKey, eventID)
}

func (ic *IdempotencyChecker) lockKey(eventID string) string {
	return fmt.Sprintf("%s:%s", ic.config.PrefixLockKey, eventID)
}

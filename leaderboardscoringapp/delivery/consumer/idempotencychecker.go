package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type Config struct {
	ProcessedKeyTTL time.Duration `koanf:"processed_key_ttl"` // 24 * time.Hour
	LockKeyTTL      time.Duration `koanf:"lock_key_ttl"`      // 5 * time.Minute

	PrefixProcessedKey string `koanf:"prefix_processed_key"` // processed_events
	PrefixLockKey      string `koanf:"prefix_lock_key"`      // lock
}

type IdempotencyChecker struct {
	redisClient *redis.Client
	config      Config
	logger      *slog.Logger
}

func NewIdempotencyChecker(client *redis.Client, config Config, logger *slog.Logger) *IdempotencyChecker {
	return &IdempotencyChecker{
		redisClient: client,
		config:      config,
		logger:      logger,
	}
}

var (
	ErrEventAlreadyProcessed = errors.New("event already processed")
	ErrEventLocked           = errors.New("event is currently locked by another processor")
)

// CheckEvent It returns specific errors if the event is a duplicate or is locked.
func (ic *IdempotencyChecker) CheckEvent(ctx context.Context, eventID string, processEventFunc func() error) error {

	if eventID == "" {
		return fmt.Errorf("invalid eventID: empty")
	}

	processedKey := ic.processedKey(eventID)
	lockKey := ic.lockKey(eventID)

	// 1. Check if the event has already been successfully processed.
	exists, err := ic.redisClient.Exists(ctx, processedKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check for processed event: %w", err)
	}
	if exists == 1 {
		return ErrEventAlreadyProcessed
	}

	// 2. Try to acquire a temporary lock.
	token := uuid.NewString()
	wasSet, sErr := ic.redisClient.SetNX(ctx, lockKey, token, ic.config.LockKeyTTL).Result()
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
	if sErr := ic.redisClient.Set(ctx, processedKey, 1, ic.config.ProcessedKeyTTL).Err(); sErr != nil {
		// This is a critical failure. The event was processed, but we couldn't mark it as such.
		ic.logger.Error(
			"CRITICAL: Failed to mark event as processed after successful execution",
			slog.String("event_id", eventID),
			slog.String("error", sErr.Error()),
		)

		return fmt.Errorf("critical: failed to mark event as processed: %w", sErr)
	}

	return nil
}

func (ic *IdempotencyChecker) releaseLock(ctx context.Context, lockKey, token string) {
	var releaseLockLua = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end
`)
	_ = releaseLockLua.Run(ctx, ic.redisClient, []string{lockKey}, token).Err()
}

func (ic *IdempotencyChecker) processedKey(eventID string) string {
	return fmt.Sprintf("%s:%s", ic.config.PrefixProcessedKey, eventID)
}

func (ic *IdempotencyChecker) lockKey(eventID string) string {
	return fmt.Sprintf("%s:%s", ic.config.PrefixLockKey, eventID)
}

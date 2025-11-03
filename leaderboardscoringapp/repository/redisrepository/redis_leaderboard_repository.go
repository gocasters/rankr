package redisrepository

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/timettl"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

// RedisLeaderboardRepository manages leaderboard using Redis Sorted Sets (ZSET)
type RedisLeaderboardRepository struct {
	client *redis.Client
}

func NewRedisLeaderboardRepository(client *redis.Client) leaderboardscoring.LeaderboardCache {
	return &RedisLeaderboardRepository{
		client: client,
	}
}

func (r *RedisLeaderboardRepository) UpsertScores(ctx context.Context, score *leaderboardscoring.UpsertScore, timeframe leaderboardscoring.Timeframe) error {
	log := logger.L()

	if score == nil {
		log.Debug("nil UpsertScore; skipping upsert")
		return nil
	}

	if len(score.Keys) == 0 || score.UserID == "" {
		log.Debug("invalid UpsertScore; skipping",
			slog.Int("keys_len", len(score.Keys)),
			slog.String("user_id", score.UserID))
		return nil
	}

	// For all_time, no expiration needed
	if timeframe == leaderboardscoring.AllTime {
		return r.upsertWithoutExpiration(ctx, score)
	}

	// Calculate expiration time for the end of the period
	expirationTime, err := timettl.CalculateEndOfPeriod(timeframe.String())
	if err != nil {
		log.Error("failed to calculate expiration time",
			slog.String("timeframe", timeframe.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("calculate expiration: %w", err)
	}

	return r.upsertWithExpiration(ctx, score, expirationTime)
}

// upsertWithoutExpiration for all_time leaderboards (no TTL)
func (r *RedisLeaderboardRepository) upsertWithoutExpiration(ctx context.Context, score *leaderboardscoring.UpsertScore) error {
	log := logger.L()
	pipe := r.client.Pipeline()

	for _, key := range score.Keys {
		pipe.ZIncrBy(ctx, key, float64(score.Score), score.UserID)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to update all_time scores",
			slog.String("user_id", score.UserID),
			slog.String("error", err.Error()))
		return fmt.Errorf("pipeline exec: %w", err)
	}

	log.Debug("successfully updated all_time scores",
		slog.String("user_id", score.UserID),
		slog.Int64("score", score.Score),
		slog.Int("keys_count", len(score.Keys)))

	return nil
}

// upsertWithExpiration updates scores and sets expiration only if not already set
func (r *RedisLeaderboardRepository) upsertWithExpiration(ctx context.Context, score *leaderboardscoring.UpsertScore, expirationTime time.Time) error {
	log := logger.L()

	// Step 1: Get TTL for all keys in batch
	pipe := r.client.Pipeline()
	ttlCmds := make([]*redis.DurationCmd, len(score.Keys))

	for i, key := range score.Keys {
		ttlCmds[i] = pipe.TTL(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to get TTLs",
			slog.String("user_id", score.UserID),
			slog.String("error", err.Error()))
		return fmt.Errorf("ttl pipeline: %w", err)
	}

	// Step 2: Increment scores and set expiration where needed
	pipe = r.client.Pipeline()

	for i, key := range score.Keys {
		// Always increment score
		pipe.ZIncrBy(ctx, key, float64(score.Score), score.UserID)

		// Set expiration only if key doesn't have TTL
		ttl := ttlCmds[i].Val()
		if ttl == -2*time.Second || ttl == -1*time.Second {
			pipe.ExpireAt(ctx, key, expirationTime)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error("failed to update scores with expiration",
			slog.String("user_id", score.UserID),
			slog.String("error", err.Error()))
		return fmt.Errorf("update pipeline: %w", err)
	}

	log.Debug("successfully updated scores with expiration",
		slog.String("user_id", score.UserID),
		slog.Int64("score", score.Score),
		slog.Int("keys_count", len(score.Keys)))

	return nil
}

func (r *RedisLeaderboardRepository) GetLeaderboard(ctx context.Context, leaderboard *leaderboardscoring.LeaderboardQuery) (leaderboardscoring.LeaderboardQueryResult, error) {
	data, err := r.client.ZRevRangeWithScores(ctx, leaderboard.Key, leaderboard.Start, leaderboard.Stop).Result()
	if err != nil {
		return leaderboardscoring.LeaderboardQueryResult{}, fmt.Errorf("zrevrange: %w", err)
	}

	rows := make([]leaderboardscoring.LeaderboardEntry, 0, len(data))
	for i, entry := range data {
		row := leaderboardscoring.LeaderboardEntry{
			Rank:   leaderboard.Start + int64(i) + 1,
			UserID: fmt.Sprintf("%v", entry.Member),
			Score:  int64(entry.Score),
		}
		rows = append(rows, row)
	}

	return leaderboardscoring.LeaderboardQueryResult{LeaderboardRows: rows}, nil
}

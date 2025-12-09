package repository

import (
	"context"
	"fmt"
	types "github.com/gocasters/rankr/type"
	"strconv"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type RedisLeaderboardRepository struct {
	client *redis.Client
}

func NewRedisLeaderboardRepository(client *redis.Client) *RedisLeaderboardRepository {
	return &RedisLeaderboardRepository{
		client: client,
	}
}

type UserScoreEntry struct {
	UserID int     `json:"user_id"`
	Score  float64 `json:"score"`
}

func (r *RedisLeaderboardRepository) GetPublicLeaderboardPaginated(ctx context.Context, projectID types.ID, page, pageSize int32) ([]UserScoreEntry, int64, error) {
	cacheKey := fmt.Sprintf("public_leaderboard:project:%d", projectID)

	start := (page - 1) * pageSize
	stop := start + pageSize - 1

	total, err := r.client.ZCard(ctx, cacheKey).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard count: %w", err)
	}

	results, err := r.client.ZRevRangeWithScores(ctx, cacheKey, int64(start), int64(stop)).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard range: %w", err)
	}

	userScores := make([]UserScoreEntry, 0, len(results))
	for _, result := range results {
		userID, ok := result.Member.(string)
		if !ok {
			continue
		}

		uid, err := strconv.Atoi(userID)
		if err != nil {
			continue
		}

		userScores = append(userScores, UserScoreEntry{
			UserID: uid,
			Score:  result.Score,
		})
	}

	return userScores, total, nil
}

func (r *RedisLeaderboardRepository) SetPublicLeaderboard(ctx context.Context, projectID types.ID, userScores map[int]float64, ttl time.Duration) error {
	log := logger.L()
	cacheKey := fmt.Sprintf("public_leaderboard:project:%d", projectID)

	pipe := r.client.Pipeline()

	pipe.Del(ctx, cacheKey)

	for userID, score := range userScores {
		pipe.ZAdd(ctx, cacheKey, redis.Z{
			Score:  score,
			Member: userID,
		})
	}

	pipe.Expire(ctx, cacheKey, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error("Failed to set public leaderboard",
			slog.String("cache_key", cacheKey),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to set leaderboard: %w", err)
	}

	log.Info("Set public leaderboard successfully",
		slog.String("cache_key", cacheKey),
		slog.Int("user_count", len(userScores)),
		slog.String("ttl", ttl.String()))

	return nil
}

package redisrepository

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type LeaderboardRepo struct {
	redisClient *redis.Client
}

func New(client *redis.Client) leaderboardscoring.Repository {
	return &LeaderboardRepo{redisClient: client}
}

func (l *LeaderboardRepo) UpsertScores(ctx context.Context, score *leaderboardscoring.UpsertScore) error {
	logger := logger.L()

	if score == nil {
		logger.Debug("nil UpsertScore; skipping upsert")
		return nil
	}

	if len(score.Keys) == 0 || score.UserID == "" {
		logger.Debug("invalid UpsertScore; skipping",
			slog.Int("keys_len", len(score.Keys)), slog.String("user_id", score.UserID))
		return nil
	}

	pipeLine := l.redisClient.Pipeline()

	for _, key := range score.Keys {
		pipeLine.ZIncrBy(ctx, key, float64(score.Score), score.UserID)
	}

	_, err := pipeLine.Exec(ctx)
	if err != nil {
		logger.Error(
			"failed to execute redisClient pipeline for updating scores",
			slog.String("user_id", score.UserID),
			slog.String("error", err.Error()),
		)
		return err
	}

	logger.Debug("successfully updated scores in redisClient pipeline", slog.String("user_id", score.UserID))

	return nil
}

func (l *LeaderboardRepo) GetLeaderboard(ctx context.Context, leaderboard *leaderboardscoring.LeaderboardQuery) (leaderboardscoring.LeaderboardQueryResult, error) {
	data, err := l.redisClient.ZRevRangeWithScores(ctx, leaderboard.Key, leaderboard.Start, leaderboard.Stop).Result()
	if err != nil {
		return leaderboardscoring.LeaderboardQueryResult{}, err
	}

	var rows = make([]leaderboardscoring.LeaderboardEntry, 0, int(leaderboard.Stop-leaderboard.Start))
	for i, entry := range data {
		var row = leaderboardscoring.LeaderboardEntry{
			Rank:   uint64(leaderboard.Start + int64(i) + 1),
			UserID: fmt.Sprintf("%v", entry.Member),
			Score:  uint64(entry.Score),
		}

		rows = append(rows, row)
	}

	return leaderboardscoring.LeaderboardQueryResult{LeaderboardRows: rows}, nil
}

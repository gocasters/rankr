package repository

import (
	"context"
	"log/slog"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type LeaderboardscoringRepo struct {
	client *redis.Client
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewLeaderboardscoringRepo(client *redis.Client, db *pgxpool.Pool, logger *slog.Logger) leaderboardscoring.Repository {
	return &LeaderboardscoringRepo{
		client: client,
		db:     db,
		logger: logger,
	}
}

func (l *LeaderboardscoringRepo) PersistContribution(ctx context.Context, event *leaderboardscoring.ContributionEvent) error {
	return nil
}

func (l *LeaderboardscoringRepo) UpdateScores(ctx context.Context, keys []string, score int, userID string) error {
	pipeLine := l.client.Pipeline()

	for _, key := range keys {
		pipeLine.ZIncrBy(ctx, key, float64(score), userID)
	}

	_, err := pipeLine.Exec(ctx)
	if err != nil {
		l.logger.Error(
			"failed to execute redis pipeline for updating scores",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return err
	}

	l.logger.Debug("successfully updated scores in redis pipeline", slog.String("user_id", userID))
	return nil
}

func (l *LeaderboardscoringRepo) AddToRedisStream(ctx context.Context, streamKey string, event *leaderboardscoring.ContributionEvent) error {
	pipeline := l.client.Pipeline()

	args := &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 1000,
		Approx: true,
		Values: event,
	}

	id, err := pipeline.XAdd(ctx, args).Result()
	if err != nil {
		return err
	}
	l.logger.Debug("successfully Added event to stream", slog.String("id", id))
	return nil
}

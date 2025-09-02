package repository

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type LeaderboardRepo struct {
	client *redis.Client
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewLeaderboardscoringRepo(client *redis.Client, db *pgxpool.Pool, logger *slog.Logger) leaderboardscoring.Repository {
	return &LeaderboardRepo{
		client: client,
		db:     db,
		logger: logger,
	}
}

func (l *LeaderboardRepo) UpsertScores(ctx context.Context, keys []string, score uint8, contributorID string) error {
	pipeLine := l.client.Pipeline()

	for _, key := range keys {
		pipeLine.ZIncrBy(ctx, key, float64(score), contributorID)
	}

	_, err := pipeLine.Exec(ctx)
	if err != nil {
		l.logger.Error(
			"failed to execute redis pipeline for updating scores",
			slog.String("user_id", contributorID),
			slog.String("error", err.Error()),
		)
		return err
	}

	l.logger.Debug("successfully updated scores in redis pipeline", slog.String("user_id", contributorID))
	return nil
}

// Enqueue TODO - Implement me
func (l *LeaderboardRepo) Enqueue(ctx context.Context, payload []byte) error {
	return leaderboardscoring.ErrNotImplemented
}

// DequeueBatch TODO - Implement me
func (l *LeaderboardRepo) DequeueBatch(ctx context.Context, batchSize int) ([][]byte, error) {
	return make([][]byte, 0), leaderboardscoring.ErrNotImplemented
}

// PersistEventBatch TODO - Implement me
func (l *LeaderboardRepo) PersistEventBatch(ctx context.Context, events []*leaderboardscoring.Event) error {
	return leaderboardscoring.ErrNotImplemented
}

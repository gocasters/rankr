package repository

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type LeaderboardRepo struct {
	client *redis.Client
	db     *pgxpool.Pool
}

func NewLeaderboardscoringRepo(client *redis.Client, db *pgxpool.Pool) leaderboardscoring.Repository {
	return &LeaderboardRepo{
		client: client,
		db:     db,
	}
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

	pipeLine := l.client.Pipeline()

	for _, key := range score.Keys {
		pipeLine.ZIncrBy(ctx, key, float64(score.Score), score.UserID)
	}

	_, err := pipeLine.Exec(ctx)
	if err != nil {
		logger.Error(
			"failed to execute redis pipeline for updating scores",
			slog.String("user_id", score.UserID),
			slog.String("error", err.Error()),
		)
		return err
	}

	logger.Debug("successfully updated scores in redis pipeline", slog.String("user_id", score.UserID))

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

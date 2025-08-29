package postgres

import (
	"github.com/gocasters/rankr/pkg/database"
	"log/slog"
)

type LeaderboardstatRepo struct {
	Logger     *slog.Logger
	PostgreSQL *database.Database
}

func NewLeaderboardstatRepo(postgresDb *database.Database, logger *slog.Logger) LeaderboardstatRepo {
	return LeaderboardstatRepo{PostgreSQL: postgresDb}
}

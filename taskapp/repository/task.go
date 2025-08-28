package repository

import (
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/taskapp/service/task"
	"log/slog"
)

type Config struct {
	CacheEnabled bool   `koanf:"cache_enabled"`
	CachePrefix  string `koanf:"cache_prefix"`
}

type TaskRepo struct {
	Config     Config
	Logger     *slog.Logger
	PostgreSQL *database.Database
	Cache      *redis.Adapter
}

func NewTaskRepo(config Config, db *database.Database, logger *slog.Logger) task.Repository {
	return &TaskRepo{
		Config:     config,
		Logger:     logger,
		PostgreSQL: db,
	}
}

package authapp

import (
	"time"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/authapp/repository"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	PostgresDB           database.Config   `koanf:"postgres_db"`
	Redis                redis.Config      `koanf:"redis"`
	Repository           repository.Config `koanf:"repository"`
	Logger               logger.Config     `koanf:"logger"`
	JWT                  JWTConfig         `koanf:"jwt"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
	PathOfMigration      string            `koanf:"path_of_migration"`
}

type JWTConfig struct {
	Secret        string        `koanf:"secret"`
	TokenDuration time.Duration `koanf:"token_duration"`
}

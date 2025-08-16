package leaderboardscoring

import (
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoring/delivery/consumer"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	PostgresDB           database.Config   `koanf:"postgres_db"`
	Redis                redis.Config      `koanf:"redis"`
	Logger               logger.Config     `koanf:"logger"`
	Consumer             consumer.Config   `koanf:"consumer"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
	PathOfMigration      string            `koanf:"path_of_migration"`
}

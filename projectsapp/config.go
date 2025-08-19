package projectsapp

import (
	"time"

	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	HTTPServer httpserver.Config `koanf:"http_server"`
	PostgresDB database.Config   `koanf:"postgres_db"`
	Logger     logger.Config     `koanf:"logger"`

	TotalShutdownTimeout time.Duration `koanf:"total_shutdown_timeout"`
	PathOfMigration      string        `koanf:"path_of_migration"`
}

package notifapp

import (
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	Logger               logger.Config     `koanf:"logger"`
	PostgresDB           database.Config   `koanf:"postgres_db"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
	PathOfMigration      string            `koanf:"path_of_migration"`
}

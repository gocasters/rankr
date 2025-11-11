package taskapp

import (
	"time"

	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/taskapp/repository"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server" json:"HTTPServer"`
	PostgresDB           database.Config   `koanf:"postgres_db" json:"postgresDB,omitempty"`
	Repository           repository.Config `koanf:"repository" json:"repository,omitempty"`
	Redis                redis.Config      `koanf:"redis" json:"redis"`
	WatermillNats        nats.Config       `koanf:"watermill_nats" json:"watermillNats,omitempty"`
	Logger               logger.Config     `koanf:"logger" json:"logger,omitempty"`
	StreamNameRawEvents  string            `koanf:"stream_name_raw_events" json:"streamNameRawEvents,omitempty"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout" json:"totalShutdownTimeout,omitempty"`
	PathOfMigration      string            `koanf:"path_of_migration" json:"pathOfMigration,omitempty"`
}

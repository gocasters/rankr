package contributorapp

import (
	"github.com/gocasters/rankr/adapter/redis"
	middleware2 "github.com/gocasters/rankr/contributorapp/delivery/http/middleware"
	"github.com/gocasters/rankr/contributorapp/repository"
	"github.com/gocasters/rankr/contributorapp/service/job"
	"github.com/gocasters/rankr/contributorapp/worker"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config       `koanf:"http_server" json:"HTTPServer"`
	PostgresDB           database.Config         `koanf:"postgres_db" json:"postgresDB,omitempty"`
	Repository           repository.Config       `koanf:"repository" json:"repository,omitempty"`
	Redis                redis.Config            `koanf:"redis" json:"redis"`
	Logger               logger.Config           `koanf:"logger" json:"logger,omitempty"`
	TotalShutdownTimeout time.Duration           `koanf:"total_shutdown_timeout" json:"totalShutdownTimeout,omitempty"`
	PathOfMigration      string                  `koanf:"path_of_migration" json:"pathOfMigration,omitempty"`
	Middleware           middleware2.Config      `koanf:"middleware" json:"middleware"`
	Job                  job.ConfigJob           `koanf:"job" json:"job"`
	Validation           job.ValidateConfig      `koanf:"validate" json:"validation"`
	Broker               repository.BrokerConfig `koanf:"broker" json:"broker"`
	Worker               worker.Config           `koanf:"worker" json:"worker"`
}

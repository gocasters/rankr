package contributorapp

import (
	"time"

	"github.com/gocasters/rankr/adapter/redis"
	middleware2 "github.com/gocasters/rankr/contributorapp/delivery/http/middleware"
	"github.com/gocasters/rankr/contributorapp/repository"
	"github.com/gocasters/rankr/contributorapp/service/job"
	"github.com/gocasters/rankr/contributorapp/worker"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	Middleware           middleware2.Config      `koanf:"middleware" json:"middleware"`
	Job                  job.ConfigJob           `koanf:"job" json:"job"`
	Validation           job.ValidateConfig      `koanf:"validate" json:"validation"`
	Broker               repository.BrokerConfig `koanf:"broker" json:"broker"`
	Worker               worker.Config           `koanf:"worker" json:"worker"`
	HTTPServer           httpserver.Config       `koanf:"http_server" json:"HTTPServer"`
	GRPCServer           grpc.ServerConfig       `koanf:"rpc_server" json:"grpcServer"`
	PostgresDB           database.Config         `koanf:"postgres_db" json:"postgresDB,omitempty"`
	Repository           repository.Config       `koanf:"repository" json:"repository,omitempty"`
	Redis                redis.Config            `koanf:"redis" json:"redis"`
	Logger               logger.Config           `koanf:"logger" json:"logger,omitempty"`
	TotalShutdownTimeout time.Duration           `koanf:"total_shutdown_timeout" json:"totalShutdownTimeout,omitempty"`
	PathOfMigration      string                  `koanf:"path_of_migration" json:"pathOfMigration,omitempty"`
}

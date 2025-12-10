package contributorapp

import (
	"time"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/repository"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server" json:"HTTPServer"`
	GRPCServer           grpc.ServerConfig `koanf:"rpc_server" json:"grpcServer"`
	PostgresDB           database.Config   `koanf:"postgres_db" json:"postgresDB,omitempty"`
	Repository           repository.Config `koanf:"repository" json:"repository,omitempty"`
	Redis                redis.Config      `koanf:"redis" json:"redis"`
	Logger               logger.Config     `koanf:"logger" json:"logger,omitempty"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout" json:"totalShutdownTimeout,omitempty"`
	PathOfMigration      string            `koanf:"path_of_migration" json:"pathOfMigration,omitempty"`
}

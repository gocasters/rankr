package leaderboardscoringapp

import (
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer"
	postgrerepository "github.com/gocasters/rankr/leaderboardscoringapp/repository/database"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config             `koanf:"http_server"`
	RPCServer            grpc.ServerConfig             `koanf:"rpc_server"`
	PostgresDB           database.Config               `koanf:"postgres_db"`
	Redis                redis.Config                  `koanf:"redis"`
	Nats                 nats.Config                   `koanf:"nats"`
	Logger               logger.Config                 `koanf:"logger"`
	Consumer             consumer.Config               `koanf:"consumer"`
	TotalShutdownTimeout time.Duration                 `koanf:"total_shutdown_timeout"`
	PathOfMigration      string                        `koanf:"path_of_migration"`
	SubscriberTopic      string                        `koanf:"subscriber_topic"`
	RetryConfig          postgrerepository.RetryConfig `koanf:"postgre_repository_retry"`
}

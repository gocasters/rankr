package leaderboardstatapp

import (
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardstatapp/delivery/scheduler"
	"github.com/gocasters/rankr/leaderboardstatapp/repository"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer            httpserver.Config `koanf:"http_server"`
	RPCServer             grpc.ServerConfig `koanf:"rpc_server"`
	PostgresDB            database.Config   `koanf:"postgres_db"`
	Repository            repository.Config `koanf:"repository"`
	Redis                 redis.Config      `koanf:"redis"`
	Logger                logger.Config     `koanf:"logger" `
	TotalShutdownTimeout  time.Duration     `koanf:"total_shutdown_timeout"`
	PathOfMigration       string            `koanf:"path_of_migration"`
	SchedulerCfg          scheduler.Config  `koanf:"scheduler_cfg"`
	LeaderboardScoringRPC grpc.ClientConfig `koanf:"leaderboard_scoring_rpc"`
}

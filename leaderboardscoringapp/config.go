package leaderboardscoringapp

import (
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/natsadapter"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer/batchprocessor"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer/rawevent"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/scheduler"
	postgrerepository "github.com/gocasters/rankr/leaderboardscoringapp/repository/database"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"time"
)

type Config struct {
	// Server configurations
	HTTPServer httpserver.Config `koanf:"http_server"`
	RPCServer  grpc.ServerConfig `koanf:"rpc_server"`

	// Scheduler configurations
	SchedulerCfg scheduler.Config `koanf:"scheduler_cfg"`

	// Data store configurations
	PostgresDB database.Config `koanf:"postgres_db"`
	Redis      redis.Config    `koanf:"redis"`

	// NATS configurations
	WatermillNats nats.Config                    `koanf:"watermill_nats"` // For raw events (push-based)
	NatsAdapter   natsadapter.Config             `koanf:"nats_adapter"`   // For processed events (native)
	PullConsumer  natsadapter.PullConsumerConfig `koanf:"pull_consumer"`  // Pull consumer config

	// Application configurations
	Logger           logger.Config                 `koanf:"logger"`
	RawEventConsumer rawevent.Config               `koanf:"raw_event_consumer"`
	BatchProcessor   batchprocessor.Config         `koanf:"batch_processor"`
	DatabaseRetry    postgrerepository.RetryConfig `koanf:"database_retry"`

	// Topics
	StreamNameRawEvents string `koanf:"stream_name_raw_events"`

	// Lifecycle
	TotalShutdownTimeout time.Duration `koanf:"total_shutdown_timeout"`

	// Database migration
	PathOfMigration string `koanf:"path_of_migration"`
}

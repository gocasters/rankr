package webhookapp

import (
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/webhookapp/schedule/insert"
	"github.com/gocasters/rankr/webhookapp/schedule/recovery"
	"time"

	"github.com/gocasters/rankr/pkg/httpserver"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	ShutDownCtxTimeout   time.Duration     `koanf:"shutdown_ctx_timeout"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
	Logger               logger.Config     `koanf:"logger"`
	NATSConfig           NATSConfig        `koanf:"nats"`
	PostgresDB           database.Config   `koanf:"postgres_db"`
	PathOfMigration      string            `koanf:"path_of_migration"`
	RedisConfig          redis.Config      `koanf:"redis"`
	InsertQueueName      string            `koanf:"insert_queue_name"`
	InsertBatchSize      int64             `koanf:"insert_batch_size"`
	RecoveryConfig       recovery.Config   `koanf:"recovery_config"`
	BulkInsertConfig     insert.Config     `koanf:"bulk_insert_config"`
}

type NATSConfig struct {
	URL                  string        `koanf:"url"`
	ConnectTimeout       time.Duration `koanf:"connect_timeout"`
	ReconnectWait        time.Duration `koanf:"reconnect_wait"`
	RetryOnFailedConnect bool          `koanf:"retry_on_failed_connect"`
	JetStreamEnabled     bool          `koanf:"jet_stream_enabled"`
	Stream               struct {
		Name     string   `koanf:"name"`
		Subjects []string `koanf:"subjects"`
	} `koanf:"stream"`
}

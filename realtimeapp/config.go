package realtimeapp

import (
	"time"

	natsadapter "github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	HTTPServer           httpserver.Config  `koanf:"http_server" json:"HTTPServer"`
	Redis                redis.Config       `koanf:"redis" json:"redis"`
	Logger               logger.Config      `koanf:"logger" json:"logger,omitempty"`
	TotalShutdownTimeout time.Duration      `koanf:"total_shutdown_timeout" json:"totalShutdownTimeout,omitempty"`
	NATS                 natsadapter.Config `koanf:"nats" json:"nats"`
	SubscribeTopics      string             `koanf:"subscribe_topics" json:"subscribe_topics"`
}

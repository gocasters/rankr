package webhookapp

import (
	"github.com/gocasters/rankr/pkg/logger"
	"time"

	"github.com/gocasters/rankr/pkg/httpserver"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	ShutDownCtxTimeout   time.Duration     `koanf:"shutdown_ctx_timeout"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
	Logger               logger.Config     `koanf:"logger"`
}

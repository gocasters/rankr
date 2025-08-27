package webhookapp

import (
	"time"

	"github.com/gocasters/rankr/pkg/httpserver"
)

type Config struct {
	Server               httpserver.Config `koanf:"server"`
	ShutDownCtxTimeout   time.Duration     `koanf:"shutdown_ctx_timeout"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
}

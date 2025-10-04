package userprofileapp

import (
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	HTTPServer httpserver.Config `koanf:"http_server"`
	Logger     logger.Config     `koanf:"logger"`
}

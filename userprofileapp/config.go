package userprofileapp

import (
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Config struct {
	HTTPServer httpserver.Config `koanf:"http_server"`
	PostgresDB database.Config   `koanf:"postgres_db"`
	Redis      redis.Config      `koanf:"redis"`
	Logger     logger.Config     `koanf:"logger"`
}

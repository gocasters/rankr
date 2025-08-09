package httpserver

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"time"
)

type Config struct {
	Port            int           `koanf:"port"`
	Cors            Cors          `koanf:"cors"`
	ShutDownTimeout time.Duration `koanf:"shutdown_context_timeout"`

	// Optional Otel middleware can be injected from outside.
	OtelMiddleware echo.MiddlewareFunc
}

type Cors struct {
	AllowOrigins []string `koanf:"allow_origins"`
}

type Server struct {
	Router *echo.Echo
	Config Config
}

func New(cfg Config) Server {
	e := echo.New()

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(
		middleware.CORSWithConfig(
			middleware.CORSConfig{
				AllowOrigins: cfg.Cors.AllowOrigins,
			},
		),
	)

	if cfg.OtelMiddleware != nil {
		e.Use(cfg.OtelMiddleware)
	}

	return Server{
		Router: e,
		Config: cfg,
	}
}

func (s Server) Start() error {
	addr := fmt.Sprintf(":%d", s.Config.Port)

	return s.Router.Start(addr)
}

func (s Server) Stop(ctx context.Context) error {
	return s.Router.Shutdown(ctx)
}

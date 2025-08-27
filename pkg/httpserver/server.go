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
	CORS            CORS          `koanf:"cors"`
	ShutdownTimeout time.Duration `koanf:"shutdown_context_timeout"`
	HideBanner      bool          `koanf:"hide_banner"`
	HidePort        bool          `koanf:"hide_port"`

	// Optional Otel middleware can be injected from outside.
	OtelMiddleware echo.MiddlewareFunc
}

type CORS struct {
	AllowOrigins []string `koanf:"allow_origins"`
}

type Server struct {
	router *echo.Echo
	config *Config
}

func New(cfg Config) (*Server, error) {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", cfg.Port)
	}

	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = DefaultShutdownTimeout
	}

	e := echo.New()

	if cfg.OtelMiddleware != nil {
		e.Use(cfg.OtelMiddleware)
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(
		middleware.CORSWithConfig(
			middleware.CORSConfig{
				AllowOrigins: cfg.CORS.AllowOrigins,
			},
		),
	)

	return &Server{
		router: e,
		config: &cfg,
	}, nil
}

func (s *Server) GetRouter() *echo.Echo {
	return s.router
}

func (s *Server) GetConfig() *Config {
	return s.config
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	s.router.HideBanner = s.config.HideBanner
	s.router.HidePort = s.config.HidePort

	return s.router.Start(addr)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.router.Shutdown(ctx)
}

package httpserver

import (
	"context"
	"fmt"
	echomiddleware "github.com/gocasters/rankr/pkg/echo_middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Host            string        `koanf:"host"`
	Port            int           `koanf:"port"`
	CORS            CORS          `koanf:"cors"`
	ShutdownTimeout time.Duration `koanf:"shutdown_context_timeout"`
	HideBanner      bool          `koanf:"hide_banner"`
	HidePort        bool          `koanf:"hide_port"`
	PublicPaths     []string      `koanf:"public_paths"`

	// Optional Otel middleware can be injected from outside.
	OtelMiddleware echo.MiddlewareFunc
}

type CORS struct {
	AllowOrigins []string `koanf:"allow_origins"`
}

type Server struct {
	router *echo.Echo
	config *Config

	requireClaimsOnce sync.Once
}

var basePublicPaths = []string{
	"/v1/login",
	"/v1/refresh-token",
	"/v1/me",
	"/ping",
	"/ping-otel",
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
	s.requireClaimsOnce.Do(func() {
		s.router.Use(
			echomiddleware.RequireUserInfo(
				echomiddleware.RequireUserInfoOptions{
					Skipper: newPublicPathSkipper(s.config.PublicPaths...),
				},
			),
		)
	})

	return s.router
}

func (s *Server) GetConfig() *Config {
	return s.config
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.router.HideBanner = s.config.HideBanner
	s.router.HidePort = s.config.HidePort

	return s.router.Start(addr)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.router.Shutdown(ctx)
}

func newPublicPathSkipper(extraPaths ...string) middleware.Skipper {
	paths := make([]string, 0, len(basePublicPaths)+len(extraPaths))
	paths = append(paths, basePublicPaths...)
	paths = append(paths, extraPaths...)
	staticPublicPathSkipper := echomiddleware.SkipExactPaths(paths...)

	return func(c echo.Context) bool {
		path := strings.TrimSuffix(c.Request().URL.Path, "/")
		return staticPublicPathSkipper(c) ||
			strings.HasSuffix(path, "/health-check") ||
			strings.HasSuffix(path, "/health_check")
	}
}

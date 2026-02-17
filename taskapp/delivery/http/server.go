package http

import (
	"context"
	"log/slog"

	echomiddleware "github.com/gocasters/rankr/pkg/echo_middleware"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
	logger     *slog.Logger
}

func New(server httpserver.Server, handler Handler, logger *slog.Logger) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
		logger:     logger,
	}
}

func (s Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}
	return nil
}

func (s Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s Server) RegisterRoutes() {
	router := s.HTTPServer.GetRouter()
	router.Use(
		echomiddleware.RequireClaimsWithConfig(
			echomiddleware.RequireClaimsConfig{
				Skipper: echomiddleware.SkipExactPaths("/health-check"),
			},
		),
	)

	v1 := router
	v1.GET("/health-check", s.healthCheck)
}

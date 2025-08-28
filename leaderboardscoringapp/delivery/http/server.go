package http

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/httpserver"
	"log/slog"
)

type Server struct {
	LeaderboardscoringSvc leaderboardscoring.Service
	HTTPServer            *httpserver.Server
	Handler               Handler
	logger                *slog.Logger
}

func New(server *httpserver.Server, handler Handler, logger *slog.Logger, leaderboardscoringSvc leaderboardscoring.Service) Server {
	return Server{
		LeaderboardscoringSvc: leaderboardscoringSvc,
		HTTPServer:            server,
		Handler:               handler,
		logger:                logger,
	}
}

func (s Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s Server) stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s Server) RegisterRoutes() {
	v1 := s.HTTPServer.GetRouter().Group("/v1")

	v1.GET("/health-check", s.healthCheck)

	leaderboardGroup := v1.Group("/leaderboard")

	leaderboardGroup.GET("/public", s.Handler.GetLeaderboard)
}

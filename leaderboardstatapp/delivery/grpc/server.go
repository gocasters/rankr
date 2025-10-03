package grpc

import (
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	"log/slog"
)

type Server struct {
	server  *grpc.RPCServer
	handler Handler
}

func New(server *grpc.RPCServer, handler Handler) Server {
	return Server{
		server:  server,
		handler: handler,
	}
}

func (s *Server) Serve() error {
	logger := logger.L()

	leaderboardstatpb.RegisterLeaderboardStatServiceServer(s.server.Server, &s.handler)

	logger.Info(
		"leaderboard-stat gRPC server started...",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	if err := s.server.Server.Serve(s.server.Listener); err != nil {
		logger.Error(
			"error in serving leaderboard-stat gRPC server",
			slog.String("error", err.Error()),
			slog.String("address", s.server.Listener.Addr().String()),
		)
		return err
	}

	logger.Info(
		"leaderboard-stat gRPC server stopped",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		logger.L().Info(
			"shutting down leaderboard-stat gRPC server",
			slog.String("address", s.server.Listener.Addr().String()),
		)

		s.server.Stop()

		logger.L().Info(
			"leaderboard-stat gRPC server shutdown completed",
			slog.String("address", s.server.Listener.Addr().String()),
		)
	}
}

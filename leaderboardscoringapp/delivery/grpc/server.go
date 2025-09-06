package grpc

import (
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/protobuf/leaderboardscoring/golang/leaderboardscoringpb"
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

	leaderboardscoringpb.RegisterLeaderboardScoringServiceServer(s.server.Server, &s.handler)

	logger.Info(
		"leaderboard-scoring gRPC server started",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	if err := s.server.Server.Serve(s.server.Listener); err != nil {
		logger.Error(
			"error in serving leaderboard-scoring gRPC server",
			slog.String("error", err.Error()),
			slog.String("address", s.server.Listener.Addr().String()),
		)
		return err
	}

	logger.Info(
		"leaderboard-scoring gRPC server stopped",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.Stop()

		logger.L().Info(
			"shutting down leaderboard-scoring gRPC server  gracefully",
			slog.String("address", s.server.Listener.Addr().String()),
		)
	}
}

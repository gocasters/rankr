package grpc

import (
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/protobuf/leaderboardscoring/golang/leaderboardscoringpb"
	"log/slog"
)

type Server struct {
	server  *grpc.RPCServer
	handler Handler
	logger  *slog.Logger
}

func New(server *grpc.RPCServer, handler Handler, logger *slog.Logger) Server {
	return Server{
		server:  server,
		handler: handler,
		logger:  logger,
	}
}

func (s *Server) Serve() error {
	leaderboardscoringpb.RegisterLeaderboardScoringServiceServer(s.server.Server, &s.handler)

	s.logger.Info(
		"leaderboard-scoring gRPC server started",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	if err := s.server.Server.Serve(s.server.Listener); err != nil {
		s.logger.Error(
			"error in serving leaderboard-scoring gRPC server",
			slog.String("error", err.Error()),
			slog.String("address", s.server.Listener.Addr().String()),
		)
		return err
	}

	s.logger.Info(
		"leaderboard-scoring gRPC server stopped",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.Stop()

		s.logger.Info(
			"shutting down leaderboard-scoring gRPC server  gracefully",
			slog.String("address", s.server.Listener.Addr().String()),
		)
	}
}

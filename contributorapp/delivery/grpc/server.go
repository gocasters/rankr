package grpc

import (
	"log/slog"

	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
	contributorpb "github.com/gocasters/rankr/protobuf/golang/contributor/v1"
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
	log := logger.L()

	contributorpb.RegisterContributorServiceServer(s.server.Server, &s.handler)

	log.Info(
		"contributor gRPC server started",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	if err := s.server.Server.Serve(s.server.Listener); err != nil {
		log.Error(
			"error in serving contributor gRPC server",
			slog.String("error", err.Error()),
			slog.String("address", s.server.Listener.Addr().String()),
		)
		return err
	}

	log.Info(
		"contributor gRPC server stopped",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		logger.L().Info(
			"shutting down contributor gRPC server",
			slog.String("address", s.server.Listener.Addr().String()),
		)

		s.server.Stop()

		logger.L().Info(
			"contributor gRPC server shutdown completed",
			slog.String("address", s.server.Listener.Addr().String()),
		)
	}
}

package grpc

import (
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
	projectpb "github.com/gocasters/rankr/protobuf/golang/project/v1"
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
	log := logger.L()

	projectpb.RegisterProjectServiceServer(s.server.Server, &s.handler)

	log.Info(
		"project gRPC server started",
		slog.String("address", s.server.Listener.Addr().String()),
	)

	if err := s.server.Server.Serve(s.server.Listener); err != nil {
		log.Error(
			"error in serving project gRPC server",
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.Stop()
		logger.L().Info("project gRPC server stopped gracefully")
	}
}

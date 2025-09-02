package grpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
)

type ServerConfig struct {
	Host        string `koanf:"host"`
	Port        int    `koanf:"port"`
	NetworkType string `koanf:"network_type"` // e.g., tcp, unix
}

type RPCServer struct {
	config   ServerConfig
	Server   *grpc.Server
	Listener net.Listener
}

func NewServer(cfg ServerConfig, logger *slog.Logger) (*RPCServer, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	listener, err := net.Listen(cfg.NetworkType, address)
	if err != nil {
		logger.Error(
			"failed to create listener for gRPC server",
			slog.String("address", address),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Configure Server Options (Interceptors)
	serverOptions := buildServerOptions(cfg, logger)

	// Create the gRPC Server Instance
	gRPCServer := grpc.NewServer(serverOptions...)

	// Register Standard Services
	// Register the gRPC Health Check service.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(gRPCServer, healthServer)
	// Manually set the health status of the main service.
	healthServer.SetServingStatus(cfg.Host, grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable Server Reflection for debugging and testing purposes.
	reflection.Register(gRPCServer)

	return &RPCServer{
		config:   cfg,
		Server:   gRPCServer,
		Listener: listener,
	}, nil
}

func (s *RPCServer) Stop() {
	s.Server.GracefulStop()
}

func buildServerOptions(cfg ServerConfig, logger *slog.Logger) []grpc.ServerOption {

	// Create a logging interceptor
	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	// Create a recovery interceptor to catch panics and return gRPC errors.
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			logger.Error("gRPC handler panicked", slog.Any("panic", p))
			return fmt.Errorf("internal server error")
		}),
	}

	// Chain the interceptors. The order is important: recovery should be first.
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptorLogger(logger), logOpts...),
		),
	}

	logger.Warn("gRPC server is starting without TLS. This is not suitable for production.")

	return opts
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		switch level {
		case logging.LevelDebug:
			l.Debug(msg, fields...)
		case logging.LevelInfo:
			l.Info(msg, fields...)
		case logging.LevelWarn:
			l.Warn(msg, fields...)
		case logging.LevelError:
			l.Error(msg, fields...)
		default:
			l.Log(ctx, slog.Level(level), msg, fields...)
		}
	})
}

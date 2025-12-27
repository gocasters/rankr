package grpc

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
)

type ServerConfig struct {
	Host          string `koanf:"host"`
	Port          int    `koanf:"port"`
	NetworkType   string `koanf:"network_type"` // e.g., tcp, unix
	TLSCertFile   string `koanf:"tls_cert_file"`
	TLSKeyFile    string `koanf:"tls_key_file"`
	AllowInsecure bool   `koanf:"allow_insecure"`
}

type RPCServer struct {
	config   ServerConfig
	Server   *grpc.Server
	Listener net.Listener
}

func NewServer(cfg ServerConfig) (*RPCServer, error) {
	logger := logger.L()
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

	logger.Info("gRPC server listening on",
		slog.String("address", address),
		slog.String("network", cfg.NetworkType))

	// Configure Server Options (Interceptors)
	serverOptions, err := buildServerOptions(cfg, logger)
	if err != nil {
		logger.Error(
			"failed to configure TLS for gRPC server; refusing to start without TLS",
			slog.String("error", err.Error()),
		)
		_ = listener.Close()
		return nil, err
	}

	// Create the gRPC Server Instance
	gRPCServer := grpc.NewServer(serverOptions...)

	// Register Standard Services
	// Register the gRPC Health Check service.
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(gRPCServer, healthServer)
	// Manually set the health status of the main service.
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

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

func buildServerOptions(cfg ServerConfig, logger *slog.Logger) ([]grpc.ServerOption, error) {
	if cfg.TLSCertFile == "" || cfg.TLSKeyFile == "" {
		if cfg.AllowInsecure {
			logger.Warn("gRPC server is running without TLS. This is not suitable for production.")
			return []grpc.ServerOption{}, nil
		}
		return nil, fmt.Errorf("tls_cert_file and tls_key_file must be configured")
	}

	creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS credentials: %w", err)
	}

	// Create a logging interceptor
	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	// Create a recovery interceptor to catch panics and return gRPC errors.
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) error {
			logger.Error("gRPC handler panicked", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal server error")
		}),
	}

	// Chain the interceptors. The order is important: recovery should be first.
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(interceptorLogger(logger), logOpts...),
		),
	}

	return opts, nil
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

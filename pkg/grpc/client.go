package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Host                 string        `koanf:"host"`
	Port                 int           `koanf:"port"`
	GRPCServiceName      string        `koanf:"grpc_service_name"`
	MaxAttempts          int           `koanf:"max_attempts"`
	InitialBackoff       time.Duration `koanf:"initial_backoff"`
	MaxBackoff           time.Duration `koanf:"max_backoff"`
	BackoffMultiplier    float64       `koanf:"backoff_multiplier"`
	RetryableStatusCodes []string      `koanf:"retryable_status_codes"`
}

type RPCClient struct {
	config ClientConfig
	Conn   *grpc.ClientConn
}

func NewClient(cfg ClientConfig, logger *slog.Logger) (*RPCClient, error) {
	target := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Build the retry policy
	retryPolicyJSON, err := buildRetryPolicyJSON(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build gRPC retry policy: %w", err)
	}

	// TODO: For production environments, transport credentials (TLS) must be re-enabled.
	// This current implementation is suitable only for local development.
	transportCreds := insecure.NewCredentials()
	logger.Warn("gRPC client is using insecure credentials. This is not suitable for production.")

	// Dial the server with all options
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithDefaultServiceConfig(retryPolicyJSON),
		grpc.WithBlock(),
	}

	timeout := cfg.InitialBackoff
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, dialOptions...)
	if err != nil {
		logger.Error(
			"failed to dial gRPC server",
			slog.String("service", cfg.GRPCServiceName),
			slog.String("target", target),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return &RPCClient{
		config: cfg,
		Conn:   conn,
	}, nil
}

func (c *RPCClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

func buildRetryPolicyJSON(cfg ClientConfig) (string, error) {

	if cfg.GRPCServiceName == "" {
		return "", fmt.Errorf("grpc_service_name must be set")
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.MaxAttempts > 5 {
		cfg.MaxAttempts = 5
	}

	type retryPolicy struct {
		MaxAttempts          int      `json:"MaxAttempts"`
		InitialBackoff       string   `json:"InitialBackoff"`
		MaxBackoff           string   `json:"MaxBackoff"`
		BackoffMultiplier    float64  `json:"BackoffMultiplier"`
		RetryableStatusCodes []string `json:"RetryableStatusCodes"`
	}

	type methodConfig struct {
		Name        []map[string]string `json:"name"`
		RetryPolicy retryPolicy         `json:"retryPolicy"`
	}

	type serviceConfig struct {
		MethodConfig []methodConfig `json:"methodConfig"`
	}

	configData := serviceConfig{
		MethodConfig: []methodConfig{
			{
				Name: []map[string]string{
					{"service": cfg.GRPCServiceName},
				},
				RetryPolicy: retryPolicy{
					MaxAttempts:          cfg.MaxAttempts,
					InitialBackoff:       fmt.Sprintf("%.1fs", cfg.InitialBackoff.Seconds()),
					MaxBackoff:           fmt.Sprintf("%.1fs", cfg.MaxBackoff.Seconds()),
					BackoffMultiplier:    cfg.BackoffMultiplier,
					RetryableStatusCodes: cfg.RetryableStatusCodes,
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(configData)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

package otel

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpcConnectionManager handles gRPC connections
type grpcConnectionManager struct {
	conn *grpc.ClientConn
}

func newGrpcConnectionManager(endpoint string) (*grpcConnectionManager, error) {
	creds := insecure.NewCredentials()

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return &grpcConnectionManager{conn: conn}, nil
}

func (g *grpcConnectionManager) GetConnection() *grpc.ClientConn {
	return g.conn
}

func (g *grpcConnectionManager) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

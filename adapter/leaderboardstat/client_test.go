// /Users/ghodsiehshahsavan/Documents/go/rankr/adapter/leaderboardstat/client_test.go
package leaderboardstat

import (
	"context"
	"net"
	"testing"
	"time"

	lbstat "github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	mygrpc "github.com/gocasters/rankr/pkg/grpc"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	types "github.com/gocasters/rankr/type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock gRPC server for testing
type mockLeaderboardStatServer struct {
	leaderboardstatpb.UnimplementedLeaderboardStatServiceServer
	getContributorStatsFunc func(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error)
}

func (m *mockLeaderboardStatServer) GetContributorStats(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
	return m.getContributorStatsFunc(ctx, req)
}

func startTestServer(t *testing.T, server leaderboardstatpb.LeaderboardStatServiceServer) (*grpc.Server, string) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s := grpc.NewServer()
	leaderboardstatpb.RegisterLeaderboardStatServiceServer(s, server)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	return s, lis.Addr().String()
}

func createTestClient(t *testing.T, serverAddr string) *Client {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	require.NoError(t, err)

	rpcClient := &mygrpc.RPCClient{Conn: conn}
	client, err := New(rpcClient)
	require.NoError(t, err)

	return client
}

func TestNew_ClientCreation(t *testing.T) {
	tests := []struct {
		name      string
		rpcClient *mygrpc.RPCClient
		wantError bool
	}{
		{
			name:      "successful client creation",
			rpcClient: &mygrpc.RPCClient{Conn: &grpc.ClientConn{}},
			wantError: false,
		},
		{
			name:      "nil rpc client",
			rpcClient: nil,
			wantError: true,
		},
		{
			name:      "nil connection",
			rpcClient: &mygrpc.RPCClient{Conn: nil},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.rpcClient)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.leaderboardStatClient)
			}
		})
	}
}

func TestClient_GetContributorStats_Success(t *testing.T) {
	mockServer := &mockLeaderboardStatServer{
		getContributorStatsFunc: func(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
			return &leaderboardstatpb.ContributorStatResponse{
				ContributorId: req.ContributorId,
				GlobalRank:    5,
				TotalScore:    1000.0,
				ProjectsScore: map[uint64]float64{1: 500.0, 2: 500.0},
			}, nil
		},
	}

	server, addr := startTestServer(t, mockServer)
	defer server.Stop()

	client := createTestClient(t, addr)
	defer client.Close()

	req := &lbstat.ContributorStatsRequest{
		ContributorID: 123,
	}

	resp, err := client.GetContributorStats(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, types.ID(123), resp.ContributorID)
	assert.Equal(t, 5, resp.GlobalRank)

	assert.Equal(t, float64(1000), resp.TotalScore)
	assert.Len(t, resp.ProjectsScore, 2)
	assert.Equal(t, float64(500), resp.ProjectsScore[types.ID(1)])
}

func TestClient_GetContributorStats_ServerError(t *testing.T) {
	mockServer := &mockLeaderboardStatServer{
		getContributorStatsFunc: func(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
			return nil, status.Error(codes.Internal, "server error")
		},
	}

	server, addr := startTestServer(t, mockServer)
	defer server.Stop()

	client := createTestClient(t, addr)
	defer client.Close()

	req := &lbstat.ContributorStatsRequest{
		ContributorID: 123,
	}

	resp, err := client.GetContributorStats(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "server error")
}

func TestClient_GetContributorStats_ContextCancellation(t *testing.T) {
	mockServer := &mockLeaderboardStatServer{
		getContributorStatsFunc: func(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
			// Simulate slow response
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return &leaderboardstatpb.ContributorStatResponse{
					ContributorId: req.ContributorId,
					GlobalRank:    1,
					TotalScore:    0.0,
					ProjectsScore: map[uint64]float64{},
				}, nil
			}
		},
	}

	server, addr := startTestServer(t, mockServer)
	defer server.Stop()

	client := createTestClient(t, addr)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &lbstat.ContributorStatsRequest{
		ContributorID: 123,
	}

	resp, err := client.GetContributorStats(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestClient_Close(t *testing.T) {
	// Create a real connection to test Close functionality
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	server := grpc.NewServer()
	leaderboardstatpb.RegisterLeaderboardStatServiceServer(server, &mockLeaderboardStatServer{
		getContributorStatsFunc: func(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
			return &leaderboardstatpb.ContributorStatResponse{}, nil
		},
	})

	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	rpcClient := &mygrpc.RPCClient{Conn: conn}
	client, err := New(rpcClient)
	require.NoError(t, err)

	// Should not panic and should close the connection
	assert.NotPanics(t, func() {
		client.Close()
	})

	// Verify connection is closed
	assert.Error(t, conn.Close()) // Closing already closed connection should error
}

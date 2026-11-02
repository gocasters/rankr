package userprofileapp

import (
	"context"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
	"github.com/stretchr/testify/require"
)

func TestConfigValidateRequiresPositiveUpstreamTimeout(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		err := (Config{Service: userprofile.Config{UpstreamTimeout: timeout}}).Validate()
		require.Error(t, err)
	}

	require.NoError(t, (Config{Service: userprofile.Config{UpstreamTimeout: time.Second}}).Validate())
}

func TestSetupSucceedsWhenUpstreamsAreUnavailable(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))
	defer logger.Close()

	cfg := Config{
		HTTPServer: httpserver.Config{Port: 18081},
		ContributorRPC: grpc.ClientConfig{
			Host: "192.0.2.1", Port: 65001, GRPCServiceName: "contributor.v1.ContributorService",
			MaxAttempts: 3, InitialBackoff: 500 * time.Millisecond, MaxBackoff: time.Second,
			BackoffMultiplier: 1.5, RetryableStatusCodes: []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"},
		},
		LeaderboardStatRPC: grpc.ClientConfig{
			Host: "192.0.2.1", Port: 65002, GRPCServiceName: "leaderboardstat.LeaderboardStatService",
			MaxAttempts: 3, InitialBackoff: 500 * time.Millisecond, MaxBackoff: time.Second,
			BackoffMultiplier: 1.5, RetryableStatusCodes: []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"},
		},
		Service: userprofile.Config{UpstreamTimeout: 5 * time.Second},
	}

	app, err := Setup(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, app.Handler)
}

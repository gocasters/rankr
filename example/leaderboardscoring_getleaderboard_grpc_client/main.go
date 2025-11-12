package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	lbclient "github.com/gocasters/rankr/adapter/leaderboardscoring"
	lbscoring "github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/grpc"
)

func main() {
	// CLI flags
	addr := flag.String("addr", "localhost:8090", "gRPC server address (e.g. localhost:8090)")
	timeframe := flag.String("timeframe", "all_time", "Leaderboard timeframe (all_time, yearly, monthly, weekly)")
	projectID := flag.String("project", "1", "Project ID (empty for global leaderboard)")
	pageSize := flag.Int("limit", 10, "Number of records to fetch")
	offset := flag.Int("offset", 0, "Offset for pagination")
	timeout := flag.Duration("timeout", 5*time.Second, "Request timeout duration")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Parse host:port
	var host string
	var port int

	parts := strings.Split(*addr, ":")
	if len(parts) != 2 {
		log.Fatalf("invalid --addr format (expected host:port): %s", *addr)
	}

	host = parts[0]
	p, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("invalid port in --addr: %v", err)
	}
	port = p

	// Setup gRPC client config
	cfg := grpc.ClientConfig{
		Host:                 host,
		Port:                 port,
		GRPCServiceName:      "leaderboardscoring.v1.LeaderboardScoringService",
		MaxAttempts:          3,
		InitialBackoff:       1 * time.Second,
		MaxBackoff:           5 * time.Second,
		BackoffMultiplier:    1.5,
		RetryableStatusCodes: []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"},
	}

	// Connect to gRPC server
	rpcClient, err := grpc.NewClient(cfg, logger)
	if err != nil {
		log.Fatalf("failed to connect gRPC server: %v", err)
	}
	defer rpcClient.Close()

	client, err := lbclient.New(rpcClient)
	if err != nil {
		log.Fatalf("failed to init leaderboard adapter client: %v", err)
	}
	defer client.Close()

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Build request
	var pidPtr *string
	if *projectID != "" {
		pidPtr = projectID
	}

	req := &lbscoring.GetLeaderboardRequest{
		Timeframe: *timeframe,
		ProjectID: pidPtr,
		PageSize:  int32(*pageSize),
		Offset:    int32(*offset),
	}

	fmt.Printf("\nFetching leaderboard from %s (timeframe=%s, project=%s)\n",
		*addr, *timeframe, func() string {
			if pidPtr != nil {
				return *pidPtr
			}
			return "global"
		}(),
	)

	// Execute request
	res, err := client.GetLeaderboard(ctx, req)
	if err != nil {
		log.Fatalf("GetLeaderboard RPC failed: %v", err)
	}

	// Display result
	fmt.Printf("\nLeaderboard: %s\n", res.Timeframe)
	if res.ProjectID != nil {
		fmt.Printf("Project ID: %s\n", *res.ProjectID)
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("%-6s %-12s %-10s\n", "Rank", "UserID", "Score")
	fmt.Println("----------------------------------------")
	for _, row := range res.LeaderboardRows {
		fmt.Printf("%-6d %-12s %-10d\n", row.Rank, row.UserID, row.Score)
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("Total rows: %d\n\n", len(res.LeaderboardRows))
}

package command

import (
	"context"
	"github.com/gocasters/rankr/adapter/leaderboardscoring"
	"github.com/gocasters/rankr/adapter/project"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardstatapp/repository"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
	"time"

	"github.com/spf13/cobra"
)

var schedulerCmd = &cobra.Command{
	Use:   "run-scheduler",
	Short: "Run the daily score calculation scheduler once manually",
	Long:  `This command runs the daily score calculation immediately for testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		runScheduler()
	},
}

func runScheduler() {
	cfg := loadAppConfig()

	if err := logger.Init(cfg.Logger); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Close()

	leaderboardLogger := logger.L()
	leaderboardLogger.Info("Running daily score calculation manually...")

	databaseConn, err := database.Connect(cfg.PostgresDB)
	if err != nil {
		leaderboardLogger.Error("failed to connect to database", "error", err)
		return
	}
	defer databaseConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create minimal service setup without servers
	redisAdapter, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		leaderboardLogger.Error("failed to initialize Redis", "err", err)
		return
	}
	cache := cachemanager.NewCacheManager(redisAdapter)

	// Initialize gRPC client for leaderboardscoring
	rpcClient, err := grpc.NewClient(cfg.LeaderboardScoringRPC, leaderboardLogger)
	if err != nil {
		leaderboardLogger.Error("failed to create RPC client!", "error", err)
		return
	}
	defer rpcClient.Close()

	lbScoringClient, err := leaderboardscoring.New(rpcClient)
	if err != nil {
		leaderboardLogger.Error("failed to create leaderboardscoring client", "error", err)
		return
	}

	projectRPCClient, err := grpc.NewClient(cfg.ProjectRPC, leaderboardLogger)
	if err != nil {
		leaderboardLogger.Error("failed to create project RPC client", "error", err)
		return
	}
	defer projectRPCClient.Close()

	projectClient, err := project.New(projectRPCClient)
	if err != nil {
		leaderboardLogger.Error("failed to create project client", "error", err)
		return
	}

	statRepo := repository.NewLeaderboardstatRepo(cfg.Repository, databaseConn)
	statValidator := leaderboardstat.NewValidator(statRepo)
	redisLeaderboardRepo := repository.NewRedisLeaderboardRepository(redisAdapter.Client())
	statSvc := leaderboardstat.NewService(statRepo, statValidator, *cache, redisLeaderboardRepo, lbScoringClient, projectClient)

	if err := statSvc.SetPublicLeaderboard(ctx); err != nil {
		leaderboardLogger.Error("SetPublicLeaderboard failed", "error", err)
		return
	}

	leaderboardLogger.Info("SetPublicLeaderboard completed successfully!")
}

func init() {
	RootCmd.AddCommand(schedulerCmd)
}

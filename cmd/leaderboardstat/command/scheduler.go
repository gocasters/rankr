package command

import (
	"context"
	"github.com/gocasters/rankr/leaderboardstatapp"
	"github.com/gocasters/rankr/pkg/database"
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
	leaderboardLogger := logger.L()
	leaderboardLogger.Info("Running daily score calculation manually...")
	cfg := loadAppConfig()

	if err := logger.Init(cfg.Logger); err != nil {
		leaderboardLogger.Error("failed to initialize logger: %v", err)
	}
	defer logger.Close()

	databaseConn, err := database.Connect(cfg.PostgresDB)
	if err != nil {
		leaderboardLogger.Error("failed to connect to database: %v", err)
	}
	defer databaseConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	app, err := leaderboardstatapp.Setup(ctx, cfg, databaseConn)
	if err != nil {
		leaderboardLogger.Error("setup failed: %v", err)
	}

	if err := app.LeaderboardstatSrv.GetDailyContributorScores(ctx); err != nil {
		leaderboardLogger.Error("daily calculation failed: %v", err)
	}

	leaderboardLogger.Info("Daily score calculation completed successfully!")
}

func init() {
	RootCmd.AddCommand(schedulerCmd)
}

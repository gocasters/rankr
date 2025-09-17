package command

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
	"log"
)

var migrateUp bool
var migrateDown bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the leaderboardscoring service",
	Long:  `This command starts the main leaderboardscoring service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

func serve() {
	// Load config
	cfg := loadAppConfig()

	// Initialize logger
	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("logger close error: %v", err)
		}
	}()
	logger := logger.L()

	// Run migrations if flags are set
	if migrateUp || migrateDown {
		if migrateUp && migrateDown {
			logger.Error("invalid flags: --migrate-up and --migrate-down cannot be used together")

			return
		}

		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			logger.Info("Running migrations up...")
			mgr.Up()
			logger.Info("Migrations up completed.")
		}
		if migrateDown {
			logger.Info("Running migrations down...")
			mgr.Down()
			logger.Info("Migrations down completed.")
		}
	}

	// TODO - Start otel tracer

	logger.Info("Starting leaderboardscoring Service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := leaderboardscoringapp.Setup(ctx, cfg)
	app.Start()

	logger.Info("Leaderboard-scoring service started")
}

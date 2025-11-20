package command

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
)

var migrateUp bool
var migrateDown bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the leaderboardstat service",
	Long:  `This command starts the main leaderboardstat service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	fmt.Println("Service is initiating........")
	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

func serve() {
	fmt.Println("Starting leaderboardstat service...")

	cfg := loadAppConfig()

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Fatalf("failed to close logger: %v", err)
		}
	}()

	leaderboardLogger := logger.L()
	// Run migrations if flags are set
	if migrateUp || migrateDown {
		if migrateUp && migrateDown {
			leaderboardLogger.Error("invalid flags: --migrate-up and --migrate-down cannot be used together")

			return
		}

		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			leaderboardLogger.Info("Running migrations up...")
			mgr.Up()
			leaderboardLogger.Info("Migrations up completed.")
		}
		if migrateDown {
			leaderboardLogger.Info("Running migrations down...")
			mgr.Down()
			leaderboardLogger.Info("Migrations down completed.")
		}
	}

	// TODO - Start otel tracer
	leaderboardLogger.Info("Starting leaderboardstat Service...")

	// Connect to the database
	databaseConn, cnErr := database.Connect(cfg.PostgresDB)

	if cnErr != nil {
		leaderboardLogger.Error("failed to connect to database", slog.Any("error", cnErr))
		//os.Exit(1)

		return
	}

	defer databaseConn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, sErr := leaderboardstatapp.Setup(ctx, cfg, databaseConn)
	if sErr != nil {
		leaderboardLogger.Error("leaderboardstat setup failed", slog.Any("error", sErr))
		return
	}

	leaderboardLogger.Info("Leaderboard-stat service started")

	// This will start the HTTP server and keep it running
	app.Start()

}

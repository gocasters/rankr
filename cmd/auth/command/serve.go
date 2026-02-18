package command

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/gocasters/rankr/authapp"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
)

var migrateUp bool
var migrateDown bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the auth service",
	Long:  `This command starts the main auth service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve()
	},
}

func init() {
	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

func serve() error {
	// Load config
	cfg := loadAppConfig()

	// Initialize logger
	if err := logger.Init(cfg.Logger); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("logger close error: %v", err)
		}
	}()
	svcLogger := logger.L()

	// Run migrations if flags are set
	if migrateUp || migrateDown {
		if migrateUp && migrateDown {
			svcLogger.Error("invalid flags: --migrate-up and --migrate-down cannot be used together")

			return fmt.Errorf("invalid flags: --migrate-up and --migrate-down cannot be used together")
		}

		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			svcLogger.Info("Running migrations up...")
			if err := mgr.Up(); err != nil {
				svcLogger.Error("failed to run migrations up", slog.Any("error", err))
				return fmt.Errorf("failed to run migrations up: %w", err)
			}
			svcLogger.Info("Migrations up completed.")
		}
		if migrateDown {
			svcLogger.Info("Running migrations down...")
			if err := mgr.Down(); err != nil {
				svcLogger.Error("failed to run migrations down", slog.Any("error", err))
				return fmt.Errorf("failed to run migrations down: %w", err)
			}
			svcLogger.Info("Migrations down completed.")
		}
	}

	// TODO - Start otel tracer

	svcLogger.Info("Starting auth Service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Connect(cfg.PostgresDB)
	if err != nil {
		svcLogger.Error("failed to connect to database", slog.Any("error", err))
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	app, err := authapp.Setup(ctx, cfg, db)
	if err != nil {
		svcLogger.Error("auth setup failed", slog.Any("error", err))
		return fmt.Errorf("auth setup failed: %w", err)
	}

	app.Start()

	svcLogger.Info("auth service started")

	return nil
}

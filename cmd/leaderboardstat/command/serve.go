package command

import (
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
	"path/filepath"
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
	fmt.Println("Starting leaderboardstat service........")
	var cfg leaderboardstatapp.Config

	// Load config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(workingDir, "deploy", "leaderboardstat", "development", "config.local.yml")
	}

	options := config.Options{
		Prefix:       "LEADERBOARDSTAT_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}
	if lErr := config.Load(options, &cfg); lErr != nil {
		log.Fatalf("Failed to load leaderboardstat config: %v", lErr)
	}

	// Initialize logger
	logger.Init(cfg.Logger)
	leaderboardLogger, _ := logger.L()
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
		leaderboardLogger.Error("fatal error occurred", "reason", "failed to connect to database", slog.Any("error", cnErr))
		return
	}
	defer databaseConn.Close()

	leaderboardLogger.Info("Leaderboard-stat service started")
}

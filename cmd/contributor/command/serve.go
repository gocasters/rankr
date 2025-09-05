package command

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gocasters/rankr/contributorapp"
	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
)

var migrateUp bool
var migrateDown bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the contributor service",
	Long:  `This command starts the main contributor service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	var cfg contributorapp.Config

	// Load config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(workingDir, "deploy", "contributor", "development", "config.yaml")
	}

	options := cfgloader.Options{
		Prefix:       "contributor_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
		Transformer:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load contributor config: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.Logger)
	contributorLogger := logger.L()

	// Run migrations if flags are set
	if migrateUp || migrateDown {
		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			contributorLogger.Info("Running migrations up...")
			mgr.Up()
			contributorLogger.Info("Migrations up completed.")
		}
		if migrateDown {
			contributorLogger.Info("Running migrations down...")
			mgr.Down()
			contributorLogger.Info("Migrations down completed.")
		}
	}

	// Start the server
	contributorLogger.Info("Starting contributor Service...")
	// Connect to the database
	conn, cnErr := database.Connect(cfg.PostgresDB)
	if cnErr != nil {
		contributorLogger.Error("Failed to connect to database", slog.Any("error", cnErr))
		os.Exit(1)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, _ := contributorapp.Setup(ctx, cfg, conn, contributorLogger)
	app.Start()
}

func init() {
	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

package command

import (
	"context"
	"github.com/gocasters/rankr/taskapp"
	"log"
	"log/slog"
	"os"
	"path/filepath"

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
	Short: "Start the task service",
	Long:  `This command starts the main task service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	var cfg taskapp.Config

	// Load config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(workingDir, "deploy", "task", "development", "config.yaml")
	}

	options := cfgloader.Options{
		Prefix:       "task_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
		Transformer:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load task config: %v", err)
	}
	// Initialize logger
	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("logger close error: %v", err)
		}
	}()
	taskLogger := logger.L()

	// Run migrations if flags are set
	if migrateUp || migrateDown {
		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			taskLogger.Info("Running migrations up...")
			mgr.Up()
			taskLogger.Info("Migrations up completed.")
		}
		if migrateDown {
			taskLogger.Info("Running migrations down...")
			mgr.Down()
			taskLogger.Info("Migrations down completed.")
		}
	}

	// Start the server
	taskLogger.Info("Starting task Service...")
	// Connect to the database
	conn, cnErr := database.Connect(cfg.PostgresDB)
	if cnErr != nil {
		taskLogger.Error("Failed to connect to database", slog.Any("error", cnErr))
		os.Exit(1)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, sErr := taskapp.Setup(ctx, cfg, conn, taskLogger)
	if sErr != nil {
		taskLogger.Error("Failed to setup taskapp", slog.Any("error", sErr))
		os.Exit(1)
	}
	app.Start()
}

func init() {
	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

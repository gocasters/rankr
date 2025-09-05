package command

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/projectapp"
	"github.com/spf13/cobra"
)

var migrateUp bool
var migrateDown bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the project service",
	Long:  `This command starts the main project service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	var cfg projectapp.Config

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(workingDir, "deploy", "project", "development", "config.yaml")
	}

	options := cfgloader.Options{
		Prefix:       "PROJECT_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load project config: %v", err)
	}

	log.Printf("Using configuration file: %s", yamlPath)
	log.Printf("Using configuration directory: %s", workingDir)
	log.Printf("config is: %+v", cfg)

	err = logger.Init(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)

	}

	projectLogger, err := logger.L()

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if migrateUp || migrateDown {
		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			projectLogger.Info("Running migrations up...")
			mgr.Up()
			projectLogger.Info("Migrations up completed.")
		}
		if migrateDown {
			projectLogger.Info("Running migrations down...")
			mgr.Down()
			projectLogger.Info("Migrations down completed.")
		}
	}

	projectLogger.Info("Starting project Service...")
	conn, cnErr := database.Connect(cfg.PostgresDB)
	if cnErr != nil {
		projectLogger.Error("fatal error occurred", "reason", "failed to connect to database", slog.Any("error", cnErr))
		os.Exit(1)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := projectapp.Setup(ctx, cfg, conn, projectLogger)
	app.Start()
}

func init() {

	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

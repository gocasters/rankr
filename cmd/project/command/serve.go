package command

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/pkg/path"
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

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "project", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "project", "development", "config.local.yml")
		}
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

	err = logger.Init(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	projectLogger := logger.L()

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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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

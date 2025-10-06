package command

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/webhookapp"
	nc "github.com/nats-io/nats.go"
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
	Short: "Start the webhook service",
	Long:  `This command starts the main webhook service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

// init registers the serve command with the root command and configures its CLI flags.
// It adds the --migrate-up and --migrate-down boolean flags (marked mutually exclusive)
// which control whether database migrations run before the server starts.
func init() {
	serveCmd.Flags().BoolVar(&migrateUp, "migrate-up", false, "Run migrations up before starting the server")
	serveCmd.Flags().BoolVar(&migrateDown, "migrate-down", false, "Run migrations down before starting the server")
	serveCmd.MarkFlagsMutuallyExclusive("migrate-up", "migrate-down")
	RootCmd.AddCommand(serveCmd)
}

// serve starts the webhook service: it loads configuration, initializes logging,
// optionally runs database migrations, connects to Postgres, sets up NATS
// (optionally JetStream), creates the message publisher, constructs the webhook
// application, and starts it.
//
// The function reads configuration from the YAML file specified by the
// CONFIG_PATH environment variable (falls back to a project-local default).
// If the package-level flags `migrateUp` or `migrateDown` are set, migrations
// will be executed before the server starts; those flags are mutually
// exclusive. On irrecoverable failures (config load, logger init, DB connect,
// NATS/JetStream or publisher initialization) the function logs the error and
// terminates the process.
func serve() {
	var cfg webhookapp.Config

	// Load config
	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(projectRoot, "deploy", "webhook", "development", "config.local.yml")
	}

	options := config.Options{
		Prefix:       "WEBHOOK_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}

	if cErr := config.Load(options, &cfg); cErr != nil {
		log.Fatalf("Failed to load webhook config: %v", cErr)
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
	lbLogger := logger.L()

	// Run migrations if flags are set
	if migrateUp || migrateDown {
		if migrateUp && migrateDown {
			lbLogger.Error("invalid flags: --migrate-up and --migrate-down cannot be used together")

			return
		}

		mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)
		if migrateUp {
			lbLogger.Info("Running migrations up...")
			mgr.Up()
			lbLogger.Info("Migrations up completed.")
		}
		if migrateDown {
			lbLogger.Info("Running migrations down...")
			mgr.Down()
			lbLogger.Info("Migrations down completed.")
		}
	}

	databaseConn, cnErr := database.Connect(cfg.PostgresDB)
	if cnErr != nil {
		lbLogger.Error("fatal error occurred", "reason", "failed to connect to database", slog.Any("error", cnErr))
		panic(cnErr)
	}
	defer databaseConn.Close()

	// Initialize nats
	marshaler := &nats.GobMarshaler{}
	natsLogger := watermill.NewStdLogger(false, false)
	natsOptions := []nc.Option{
		nc.RetryOnFailedConnect(cfg.NATSConfig.RetryOnFailedConnect),
		nc.Timeout(cfg.NATSConfig.ConnectTimeout),
		nc.ReconnectWait(cfg.NATSConfig.ReconnectWait),
	}

	// connect to external NATS (container service name "nats")
	natsURL := cfg.NATSConfig.URL
	if natsURL == "" {
		natsURL = "nats://supersecret@localhost:4222" // default
	}

	natsDB, ncErr := nc.Connect(natsURL,
		nc.Timeout(cfg.NATSConfig.ConnectTimeout),
	)
	if ncErr != nil {
		lbLogger.Error("NATS not available", slog.String("error", ncErr.Error()))
		panic(ncErr)
	}
	defer natsDB.Close()

	if cfg.NATSConfig.JetStreamEnabled {
		js, jsErr := natsDB.JetStream()
		if jsErr != nil {
			lbLogger.Error("failed to init JetStream context: ", slog.String("error", jsErr.Error()))
			panic(jsErr)
		}

		_, siErr := js.StreamInfo(cfg.NATSConfig.Stream.Name)
		if siErr != nil {
			_, saErr := js.AddStream(&nc.StreamConfig{
				Name:     cfg.NATSConfig.Stream.Name,
				Subjects: cfg.NATSConfig.Stream.Subjects, // adjust to your domain naming
				Storage:  nc.FileStorage,                 // persist on disk
			})
			if saErr != nil {
				lbLogger.Error("Failed to create JetStream stream: ", slog.String("error", saErr.Error()))
				panic(saErr)
			}
		}
	}

	jsConfig := nats.JetStreamConfig{Disabled: !cfg.NATSConfig.JetStreamEnabled}

	publisher, pErr := nats.NewPublisher(
		nats.PublisherConfig{
			URL:         natsURL,
			NatsOptions: natsOptions,
			Marshaler:   marshaler,
			JetStream:   jsConfig,
		},
		natsLogger,
	)
	if pErr != nil {
		lbLogger.Error("Failed to start publisher", slog.String("error", pErr.Error()))
		panic(pErr)
	} else {
		lbLogger.Info("publisher started on " + natsURL)
	}
	defer func() {
		if pcErr := publisher.Close(); pcErr != nil {
			lbLogger.Warn("publisher close error", slog.Any("error", pcErr))
		}
	}()

	ctx := context.Background()
	adapter, adErr := redis.New(ctx, cfg.RedisConfig)
	if adErr != nil {
		lbLogger.Error("Failed to start redis adapter", slog.String("error", adErr.Error()))
	}

	app := webhookapp.Setup(cfg, databaseConn, publisher, adapter)
	app.Start()
}

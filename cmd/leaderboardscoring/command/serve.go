package command

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	wnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/gocasters/rankr/leaderboardscoringapp"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"time"

	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
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
	var cfg leaderboardscoringapp.Config

	// Load config
	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(projectRoot, "deploy", "leaderboardscoring", "development", "config.local.yml")
	}

	options := config.Options{
		Prefix:       "LEADERBOARDSCORING_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}

	if cErr := config.Load(options, &cfg); cErr != nil {
		log.Fatalf("Failed to load leaderboardscoring config: %v", cErr)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
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

	// TODO - Start otel tracer

	lbLogger.Info("Starting leaderboardscoring Service...")

	// Connect to the database
	databaseConn, cnErr := database.Connect(cfg.PostgresDB)
	if cnErr != nil {
		lbLogger.Error("fatal error occurred", "reason", "failed to connect to database", slog.Any("error", cnErr))

		return
	}
	defer databaseConn.Close()

	// TODO - When the NATS adapter is created, make sure to use that adapter and its Subscriber method to create the subscriber.
	wmLogger := watermill.NewStdLogger(true, true)
	subscriber, sErr := newJetStreamSubscriber("nats://127.0.0.1:4222", wmLogger)
	if sErr != nil {
		lbLogger.Error("Failed to create Subscriber", slog.String("error", sErr.Error()))
		panic(sErr)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := leaderboardscoringapp.Setup(ctx, cfg, lbLogger, subscriber, databaseConn, wmLogger)
	app.Start()

	lbLogger.Info("Leaderboard-scoring service started")
}

// TODO - When the NATS adapter is created, make sure to use that adapter and its Subscriber method to create the subscriber.
func newJetStreamSubscriber(natsURL string, logger watermill.LoggerAdapter) (message.Subscriber, error) {

	opts := []nats.Option{
		nats.Name("rankr-leaderboard-consumer"),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(1 * time.Second),
		nats.Timeout(10 * time.Second),
	}

	jsCfg := wnats.JetStreamConfig{
		Disabled:      false,
		AutoProvision: true,
		SubscribeOptions: []nats.SubOpt{
			nats.DeliverAll(),
			nats.AckExplicit(),
			nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
		},
		AckAsync:      false,
		DurablePrefix: "lbscoring",
	}

	marshaler := &wnats.NATSMarshaler{}

	return wnats.NewSubscriber(wnats.SubscriberConfig{
		URL:            natsURL,
		NatsOptions:    opts,
		Unmarshaler:    marshaler,
		AckWaitTimeout: 30 * time.Second,
		JetStream:      jsCfg,
	}, logger)
}

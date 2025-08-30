package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/webhookapp"
	nc "github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the webhook service",
	Long:  `This command starts the main webhook service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

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
	lbLogger, lErr := logger.L()
	if lErr != nil {
		log.Fatalf("Failed to Initialize logger: %v", lErr)
	}

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
		return
	}
	defer natsDB.Close()

	js, jsErr := natsDB.JetStream()
	if jsErr != nil {
		lbLogger.Error("failed to init JetStream context: ", slog.String("error", jsErr.Error()))
		return
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
		return
	} else {
		lbLogger.Info("Publisher started on " + natsURL)
	}

	app := webhookapp.Setup(cfg, lbLogger, publisher)
	app.Start()
}

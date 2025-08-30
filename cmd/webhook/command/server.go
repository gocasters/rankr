package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/webhookapp"
	"github.com/nats-io/nats-server/v2/server"
	nc "github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"
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

	svr, nsErr := server.NewServer(&server.Options{
		Host: "0.0.0.0",
		Port: 42222,
	})

	if nsErr != nil {
		lbLogger.Error("Failed to create nats server", slog.String("error", nsErr.Error()))
	}

	svr.Start()
	defer svr.Shutdown()

	marshaler := &nats.GobMarshaler{}
	natsLogger := watermill.NewStdLogger(false, false)
	natsOptions := []nc.Option{
		nc.RetryOnFailedConnect(true),
		nc.Timeout(30 * time.Second),
		nc.ReconnectWait(1 * time.Second),
	}

	jsConfig := nats.JetStreamConfig{Disabled: true}

	publisher, pErr := nats.NewPublisher(
		nats.PublisherConfig{
			URL:         svr.ClientURL(),
			NatsOptions: natsOptions,
			Marshaler:   marshaler,
			JetStream:   jsConfig,
		},
		natsLogger,
	)
	if pErr != nil {
		lbLogger.Error("Failed to start publisher", slog.String("error", pErr.Error()))
	} else {
		lbLogger.Info("Publisher started on " + svr.ClientURL())
	}

	app := webhookapp.Setup(cfg, lbLogger, publisher)
	app.Start()
}

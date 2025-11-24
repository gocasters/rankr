package command

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/realtimeapp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the realtime service",
	Long:  `This command starts the main realtime service.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	var cfg realtimeapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "realtime", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "realtime", "development", "config.local.yml")
		}
	}

	options := cfgloader.Options{
		Prefix:       "realtime_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
		Transformer:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load realtime config: %v", err)
	}

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("logger close error: %v", err)
		}
	}()
	realtimeLogger := logger.L()

	realtimeLogger.Info("Starting realtime Service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, setupErr := realtimeapp.Setup(ctx, cfg, realtimeLogger)
	if setupErr != nil {
		realtimeLogger.Error("Failed to setup application", slog.Any("error", setupErr))
		os.Exit(1)
	}

	app.Start()
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

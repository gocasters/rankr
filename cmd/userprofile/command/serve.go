package command

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/userprofileapp"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start userprofile service",
	Long:  "This command starts the main userprofile service.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("User profile service initializing...")

		serve()
	},
}

func init() {
	RootCmd.AddCommand(ServeCmd)
}

func serve() {
	cfg := loadAppConfig()

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	defer func() {
		if err := logger.Close(); err != nil {
			log.Fatalf("failed to cloes logger: %v", err)
		}
	}()

	userProfileAppLogger := logger.L()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := userprofileapp.Setup(ctx, cfg)
	if err != nil {
		userProfileAppLogger.Error(fmt.Sprintf("failed user profile app setup: %v", err))
		return
	}

	userProfileAppLogger.Info("UserProfile service started")

	app.Start()
}

func loadAppConfig() userprofileapp.Config {
	var cfg userprofileapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "userprofile", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "userprofile", "development", "config.local.yml")
		}
	}

	log.Printf("Loading user profile service config from: %s", yamlPath)

	option := config.Options{
		Prefix:       "USERPROFILE_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
		Transformer:  nil,
	}

	if err := config.Load(option, &cfg); err != nil {
		log.Fatalf("failed to load user profile app config: %v", err)
	}

	return cfg
}

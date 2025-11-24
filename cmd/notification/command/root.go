package command

import (
	"github.com/gocasters/rankr/notifapp"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var RootCmd = &cobra.Command{
	Use:   "notification_service",
	Short: "A CLI for notification service",
	Long: `notification Service CLI is a tool to manage and run 
the notification service, including migrations and server startup.`,
}

func loadAppConfig() notifapp.Config {
	var cfg notifapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "notification", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "notification", "development", "config.local.yml")
		}
	}

	options := config.Options{
		Prefix:       "STAT_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
		Transformer:  nil,
	}

	if err := config.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load notification config: %v", err)
	}

	return cfg
}

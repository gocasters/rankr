package command

import (
	"github.com/gocasters/rankr/leaderboardstatapp"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var RootCmd = &cobra.Command{
	Use:   "leaderboardstat_service",
	Short: "A CLI for leaderboardstat service",
	Long: `leaderboardstat Service CLI is a tool to manage and run 
the leaderboardstat service, including migrations and server startup.`,
}

func loadAppConfig() leaderboardstatapp.Config {
	var cfg leaderboardstatapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")

	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "leaderboardstat", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "leaderboardstat", "development", "config.local.yml")
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
		log.Fatalf("Failed to load leaderboardstat config: %v", err)
	}

	return cfg
}

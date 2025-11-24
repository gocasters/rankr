package command

import (
	"github.com/gocasters/rankr/leaderboardstatapp"
	"github.com/gocasters/rankr/pkg/config"
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

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")

	if yamlPath == "" {
		yamlPath = filepath.Join(workingDir, "deploy", "leaderboardstat", "development", "config.yml")
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

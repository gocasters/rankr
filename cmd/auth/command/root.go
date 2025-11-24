package command

import (
	"github.com/gocasters/rankr/authapp"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

func loadAppConfig() authapp.Config {
	var cfg authapp.Config

	yamlPath := os.Getenv("CONFIG_PATH")

	// If not set, fall back to finding the project root (ideal for local development).
	if yamlPath == "" {
		log.Println("CONFIG_PATH not set, finding project root for local dev...")
		projectRoot, err := path.PathProjectRoot()
		if err != nil {
			log.Fatalf("CONFIG_PATH not set, and failed to find project root: %v", err)
		}
		// Use the SAME config file as the 'serve' command.
		yamlPath = filepath.Join(projectRoot, "deploy", "auth", "development", "config.yml")
	}

	log.Printf("Loading configuration from: %s", yamlPath)

	options := config.Options{
		Prefix:       "auth_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}
	if err := config.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load authapp config: %v", err)
	}

	return cfg
}

var RootCmd = &cobra.Command{
	Use:   "auth_service",
	Short: "A CLI for auth service",
	Long: `auth Service CLI is a tool to manage and run 
the auth service, including migrations and server startup.`,
}

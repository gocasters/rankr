package command

import (
	"fmt"
	"github.com/gocasters/rankr/contributorapp"
	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var up bool
var down bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `This command runs the database migrations for the contributor service.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

func migrate() {
	var cfg contributorapp.Config

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(workingDir, "contributorapp", "repository", "dbconfig.yml")
	}
	// to run migrations when you want to run contributor service locally
	if path := os.Getenv("DBCONFIG_OVERRIDE_PATH"); path != "" {
		yamlPath = path
		log.Printf("Using override config: %s", yamlPath)
	} else {
		log.Printf("Using default config: %s", yamlPath)
	}

	options := cfgloader.Options{
		Prefix:       "contributor_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
		Transformer:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)

	if up {
		if err := mgr.Up(); err != nil {
			fmt.Printf("failed to migrate up: %v", err)
		}
	} else if down {
		mgr.Down()
	} else {
		log.Println("Please specify a migration direction with --up or --down")
		os.Exit(2)
	}
}

func init() {
	migrateCmd.Flags().BoolVar(&up, "up", false, "Run migrations up")
	migrateCmd.Flags().BoolVar(&down, "down", false, "Run migrations down")
	migrateCmd.MarkFlagsMutuallyExclusive("up", "down")
	RootCmd.AddCommand(migrateCmd)
}

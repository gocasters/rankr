package command

import (
	"log"
	"os"
	"path/filepath"

	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/projectapp"
	"github.com/spf13/cobra"
)

var up bool
var down bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `This command runs the database migrations for the project service.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

func migrate() {
	var cfg projectapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "project", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "project", "development", "config.local.yml")
		}
	}

	log.Printf("Using config: %s", yamlPath)

	options := cfgloader.Options{
		Prefix:       "PROJECT_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)

	if up {
		log.Println("Running migrations up...")
		mgr.Up()
		log.Println("Migrations up completed.")
	} else if down {
		log.Println("Running migrations down...")
		mgr.Down()
		log.Println("Migrations down completed.")
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

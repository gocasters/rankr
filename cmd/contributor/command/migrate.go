package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gocasters/rankr/contributorapp"
	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
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

	yamlPath := filepath.Join(workingDir, "contributorapp", "repository", "dbconfig.yml")

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
		mgr.Up()
	} else if down {
		mgr.Down()
	} else {
		log.Println("Please specify a migration direction with --up or --down")
	}
}

func init() {
	migrateCmd.Flags().BoolVar(&up, "up", false, "Run migrations up")
	migrateCmd.Flags().BoolVar(&down, "down", false, "Run migrations down")
	RootCmd.AddCommand(migrateCmd)
}

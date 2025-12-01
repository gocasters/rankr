package command

import (
	"github.com/gocasters/rankr/taskapp"
	"log"
	"os"
	"path/filepath"

	cfgloader "github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/spf13/cobra"
)

var up bool
var down bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `This command runs the database migrations for the task service.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

func migrate() {
	var cfg taskapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		defaultConfig := filepath.Join(projectRoot, "deploy", "task", "development", "config.yml")
		if _, err := os.Stat(defaultConfig); err == nil {
			yamlPath = defaultConfig
		} else {
			yamlPath = filepath.Join(projectRoot, "deploy", "task", "development", "config.local.yml")
		}
	}

	// to run migrations when you want to run task service locally
	if overridePath := os.Getenv("DBCONFIG_OVERRIDE_PATH"); overridePath != "" {
		yamlPath = overridePath
		log.Printf("Using override config: %s", yamlPath)
	} else {
		log.Printf("Using default config: %s", yamlPath)
	}

	options := cfgloader.Options{
		Prefix:       "task_",
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
		os.Exit(2)
	}
}

func init() {
	migrateCmd.Flags().BoolVar(&up, "up", false, "Run migrations up")
	migrateCmd.Flags().BoolVar(&down, "down", false, "Run migrations down")
	migrateCmd.MarkFlagsMutuallyExclusive("up", "down")
	RootCmd.AddCommand(migrateCmd)
}

package command

import (
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/gocasters/rankr/webhookapp"
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
	Long:  `This command runs the database migrations for the webhook service.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

// init registers the migrate command with the root command and defines its CLI flags.
// It adds boolean flags `--up` and `--down` to control migration direction.
func init() {
	migrateCmd.Flags().BoolVar(&up, "up", false, "Run migrations up")
	migrateCmd.Flags().BoolVar(&down, "down", false, "Run migrations down")
	RootCmd.AddCommand(migrateCmd)
}

// migrate runs database migrations for the webhook service.
// 
// It loads configuration into a webhookapp.Config from a YAML file located by
// default at "<workingDir>/webhookapp/repository/dbconfig.local.yml" unless the
// path is overridden via the DBCONFIG_OVERRIDE_PATH environment variable.
// The loaded configuration is used to construct a migrator (from cfg.PostgresDB
// and cfg.PathOfMigration). The command honors the package-level boolean flags
// `up` and `down`: `--up` runs migrations up, `--down` runs migrations down.
// If both flags are set the function logs a fatal error and exits; if neither
// is set it logs a message asking the user to specify a direction. Errors
// obtaining the working directory or loading the config are logged fatally.
func migrate() {
	var cfg webhookapp.Config

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	yamlPath := filepath.Join(workingDir, "webhookapp", "repository", "dbconfig.local.yml")

	// to run migrations when you want to run webhook service locally
	if path := os.Getenv("DBCONFIG_OVERRIDE_PATH"); path != "" {
		yamlPath = path
		log.Printf("Using override config: %s", yamlPath)
	} else {
		log.Printf("Using default config: %s", yamlPath)
	}

	options := config.Options{
		Prefix:       "WEBHOOK_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}

	if lErr := config.Load(options, &cfg); lErr != nil {
		log.Fatalf("Failed to load config: %v", lErr)
	}

	mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)

	if up && down {
		log.Fatalf("Flags --up and --down are mutually exclusive")
	}

	if up {
		mgr.Up()
	} else if down {
		mgr.Down()
	} else {
		log.Println("Please specify a migration direction with --up or --down")
	}
}

package command

import (
	"github.com/gocasters/rankr/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/config"
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
	Long:  `This command runs the database migrations for the leaderboardscoring service.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

func init() {
	migrateCmd.Flags().BoolVar(&up, "up", false, "Run migrations up")
	migrateCmd.Flags().BoolVar(&down, "down", false, "Run migrations down")
	RootCmd.AddCommand(migrateCmd)
}

func migrate() {
	var cfg leaderboardscoring.Config

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	yamlPath := filepath.Join(workingDir, "leaderboardscoring", "repository", "dbconfig.yml")

	// to run migrations when you want to run leaderboardscoring service locally
	if path := os.Getenv("DBCONFIG_OVERRIDE_PATH"); path != "" {
		yamlPath = path
		log.Printf("Using override config: %s", yamlPath)
	} else {
		log.Printf("Using default config: %s", yamlPath)
	}

	options := config.Options{
		Prefix:       "LEADERBOARDSCORING_",
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

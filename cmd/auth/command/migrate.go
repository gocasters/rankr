package command

import (
	"github.com/gocasters/rankr/pkg/migrator"
	"github.com/spf13/cobra"
	"log"
)

var up bool
var down bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `This command runs the database migrations for the auth service.`,
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
	cfg := loadAppConfig()

	mgr := migrator.New(cfg.PostgresDB, cfg.PathOfMigration)

	if up && down {
		log.Fatalf("Flags --up and --down are mutually exclusive")
	}

	if up {
		if err := mgr.Up(); err != nil {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
	} else if down {
		if err := mgr.Down(); err != nil {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
	} else {
		log.Println("Please specify a migration direction with --up or --down")
	}
}

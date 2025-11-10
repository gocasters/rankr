package command

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/notifapp"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/spf13/cobra"
	"log"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start notification service",
	Long:  "This command starts the main notification service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing Notification service ...")
		serve()
	},
}

func init() {
	RootCmd.AddCommand(ServeCmd)
}

func serve() {
	cfg := loadAppConfig()

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	defer func() {
		if err := logger.Close(); err != nil {
			log.Fatalf("failed to close logger: %v", err)
		}
	}()

	notificationAppLogger := logger.L()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := notifapp.Setup(ctx, cfg)
	if err != nil {
		notificationAppLogger.Error(fmt.Sprintf("failed notifapp setup: %v", err))
		return
	}

	notificationAppLogger.Info("Notifapp service started")

	app.Start()
}

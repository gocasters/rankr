package webhookapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/adapter/webhook/github"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/webhookapp/repository"
	"github.com/gocasters/rankr/webhookapp/schedule/insert"
	"github.com/gocasters/rankr/webhookapp/schedule/recovery"
	"github.com/gocasters/rankr/webhookapp/service/delivery"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/webhookapp/delivery/http"
)

type Application struct {
	HTTPServer          http.Server
	EventRepo           delivery.EventRepository
	Config              Config
	RecoveryScheduler   *recovery.LostDeliveriesScheduler
	BulkInsertScheduler *insert.BulkInsertScheduler
}

// Setup builds and returns an Application configured with the provided config, logger,
// database connection, and message publisher.
//
// It creates a webhook event repository from conn.Pool, constructs the HTTP server
// and delivery layer wired to a service using that repository and the publisher,
// and returns an Application with HTTPServer, EventRepo, Logger, and Config populated.
// Note: this function panics if initializing the HTTP service (httpserver.New) fails.
func Setup(config Config, conn *database.Database, pub message.Publisher, redisAdapter *redis.Adapter) Application {
	eventDurableRepo := repository.NewWebhookDurableRepository(redisAdapter)
	eventRepo := repository.NewWebhookRepository(conn.Pool)
	httpService, err := httpserver.New(config.HTTPServer)
	if err != nil {
		panic(err)
	}
	deliveryService := delivery.New(&eventRepo, pub, &eventDurableRepo, config.InsertQueueName, config.InsertBatchSize)
	appHttpServer := http.New(
		httpService,
		http.NewHandler(),
		deliveryService,
	)

	recoveryScheduler := recovery.NewSchedulerService(
		config.RecoveryConfig,
		*deliveryService,
		github.NewGitHubClient(),
	)

	bulkInsertScheduler := insert.NewSchedulerService(config.BulkInsertConfig, *deliveryService)

	return Application{
		HTTPServer:          appHttpServer,
		EventRepo:           &eventRepo,
		Config:              config,
		RecoveryScheduler:   recoveryScheduler,
		BulkInsertScheduler: bulkInsertScheduler,
	}
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	done := make(chan bool)

	startServers(app, &wg)
	startBulkInsertScheduler(app, done, &wg)
	startRecoveryScheduler(app, done, &wg)
	<-ctx.Done()
	logger.L().Info("✅ Shutdown signal received...")

	close(done)

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		logger.L().Info("✅ Servers shut down gracefully")
	} else {
		logger.L().Warn("❌ Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	logger.L().Info("✅ webhook-app stopped")

}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.L().Info(fmt.Sprintf("✅ HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			logger.L().Error(fmt.Sprintf("❌ error in HTTP server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		logger.L().Info(fmt.Sprintf("✅ HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	logger.L().Info("✅ Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		logger.L().Info("✅ All servers have been shut down successfully.")
	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	logger.L().Info(fmt.Sprintf("✅ Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))
	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.ShutDownCtxTimeout)
	defer httpCancel()
	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		logger.L().Error(fmt.Sprintf("❌ HTTP server graceful shutdown failed: %v", err))
	}

	logger.L().Info("✅ HTTP server shut down successfully.")
}

func startRecoveryScheduler(app Application, done <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	logger.L().Info("🚀 Starting recovery scheduler",
		slog.Int("interval_seconds", app.Config.RecoveryConfig.RecoveryLostDeliveriesIntervalInSeconds),
		slog.Int("webhooks_count", len(app.Config.RecoveryConfig.Webhooks)))

	go app.RecoveryScheduler.Start(done, wg)
}

func startBulkInsertScheduler(app Application, done <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	logger.L().Info("🚀 Starting recovery scheduler",
		slog.Int("interval_seconds", app.Config.BulkInsertConfig.BulkInsertIntervalInSeconds),
	)

	go app.RecoveryScheduler.Start(done, wg)
}

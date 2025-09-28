package webhookapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/webhookapp/repository"
	"github.com/gocasters/rankr/webhookapp/repository/rawevent"
	"github.com/gocasters/rankr/webhookapp/repository/serializedevent"
	"github.com/gocasters/rankr/webhookapp/service/publishevent"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/webhookapp/delivery/http"

	"github.com/go-co-op/gocron"
)

type Application struct {
	HTTPServer http.Server
	EventRepo  publishevent.EventRepository
	//EventRepo  service.EventRepository
	Logger *slog.Logger
	Config Config
	Sch    *gocron.Scheduler
}

// Setup builds and returns an Application configured with the provided config, logger,
// database connection, and message publisher.
//
// It creates a webhook event repository from conn.Pool, constructs the HTTP server
// and delivery layer wired to a service using that repository and the publisher,
// and returns an Application with HTTPServer, EventRepo, Logger, and Config populated.
// Note: this function panics if initializing the HTTP service (httpserver.New) fails.
func Setup(config Config, logger *slog.Logger, conn *database.Database, pub message.Publisher, redisAdapter *redis.Adapter) Application {
	eventDurableRepo := repository.NewWebhookDurableRepository(redisAdapter)
	eventRepo := serializedevent.NewWebhookRepository(conn.Pool)
	rawEventRepo := rawevent.NewRawWebhookRepository(conn.Pool)

	httpService, err := httpserver.New(config.HTTPServer)
	if err != nil {
		panic(err)
	}
	appHttpServer := http.New(
		httpService,
		http.NewHandler(logger),
		publishevent.New(eventRepo, rawEventRepo, pub, &eventDurableRepo, config.InsertQueueName, config.InsertBatchSize),
	)

	return Application{
		HTTPServer: appHttpServer,
		EventRepo:  eventRepo,
		Logger:     logger,
		Config:     config,
		Sch:        gocron.NewScheduler(time.UTC),
	}
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	runSchedules(app, ctx, &wg)
	<-ctx.Done()
	app.Logger.Info("✅ Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		app.Logger.Info("✅ Servers shut down gracefully")
	} else {
		app.Logger.Warn("❌ Shutdown timed out, exiting application")
		os.Exit(1)
	}
	app.Logger.Info("✅ webhook-app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("✅ HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("❌ error in HTTP server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		app.Logger.Info(fmt.Sprintf("✅ HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("✅ Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		app.Logger.Info("✅ All servers have been shut down successfully.")
	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info(fmt.Sprintf("✅ Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))
	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.ShutDownCtxTimeout)
	defer httpCancel()
	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		app.Logger.Error(fmt.Sprintf("❌ HTTP server graceful shutdown failed: %v", err))
	}

	app.Logger.Info("✅ HTTP server shut down successfully.")
}

func (app Application) BulkInsertEvents() {
	defer func() {
		if r := recover(); r != nil {
			app.Logger.Error("Panic recovered in bulk insert job",
				slog.Any("recovery", r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app.Logger.Info("🔄 Starting bulk insert processing...")

	err := app.HTTPServer.Service.ProcessBatch(ctx)
	if err != nil {
		app.Logger.Error("Bulk insert job failed",
			slog.Any("error", err),
			slog.String("note", "job will retry on next schedule"))
		return
	}

	app.Logger.Info("✅ Bulk insert job completed successfully")
}

func runSchedules(app Application, ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		app.Sch.SingletonMode()
		app.Sch.WaitForScheduleAll()

		_, err := app.Sch.Every(app.Config.BulkInsertEventsIntervalInSeconds).Second().Do(app.BulkInsertEvents)
		if err != nil {
			app.Logger.Error("Failed to schedule bulk insert job", slog.Any("error", err))
			return
		}

		app.Logger.Info("🚀 Scheduler started",
			slog.Int("interval_seconds", app.Config.BulkInsertEventsIntervalInSeconds))

		app.Sch.StartAsync()

		<-ctx.Done()
		app.Logger.Info("Shutdown signal received, stopping scheduler...")

		// Stop scheduler gracefully
		app.Sch.Stop()
		app.Logger.Info("✅ Scheduler stopped gracefully")
	}()
}

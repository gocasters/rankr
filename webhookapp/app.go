package webhookapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/webhookapp/repository"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/webhookapp/delivery/http"
	"github.com/gocasters/rankr/webhookapp/service"
)

type Application struct {
	HTTPServer http.Server
	EventRepo  service.EventRepository
	Logger     *slog.Logger
	Config     Config
}

func Setup(config Config, logger *slog.Logger, conn *database.Database, pub message.Publisher) Application {
	eventRepo := repository.NewWebhookRepository(conn.Pool)
	httpService, err := httpserver.New(config.HTTPServer)
	if err != nil {
		panic(err)
	}
	appHttpServer := http.New(
		httpService,
		http.NewHandler(logger),
		service.New(eventRepo, pub),
	)

	return Application{
		HTTPServer: appHttpServer,
		EventRepo:  eventRepo,
		Logger:     logger,
		Config:     config,
	}
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
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

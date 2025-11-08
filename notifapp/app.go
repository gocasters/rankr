package notifapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/notifapp/delivery/http"
	"github.com/gocasters/rankr/notifapp/repository"
	"github.com/gocasters/rankr/notifapp/service/notification"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func notifLogger() *slog.Logger {
	return logger.L()
}

type Application struct {
	Repo    repository.Repository
	Service notification.Service
	Handler http.Handler
	Server  http.Server
	Config  Config
}

func Setup(cfg Config) (Application, error) {
	postgresConn, err := database.Connect(cfg.PostgresDB)
	if err != nil {
		return Application{}, err
	}

	repo := repository.New(postgresConn)
	validate := notification.NewValidation()
	service := notification.NewService(repo, validate)
	handler := http.NewHandler(service)

	httpServer, err := httpserver.New(cfg.HTTPServer)
	if err != nil {
		return Application{}, err
	}

	server := http.NewServer(httpServer, handler)

	notifLogger().Info("Done setup notifapp")

	return Application{
		Repo:    repo,
		Service: service,
		Handler: handler,
		Server:  server,
		Config:  cfg,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	<-ctx.Done()

	notifLogger().Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		notifLogger().Info("Servers shutdown gracefully")
	} else {
		notifLogger().Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	notifLogger().Info("notifapp stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		notifLogger().Info(fmt.Sprintf("HTTP server starting on port %d", app.Config.HTTPServer.Port))
		if err := app.Server.Serve(); err != nil {
			notifLogger().Error(fmt.Sprintf("error listen and serve HTTP server on port %d",
				app.Config.HTTPServer.Port), "error", err)
		}

		notifLogger().Info(fmt.Sprintf("HTTP server stopped on port %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	notifLogger().Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServe(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		notifLogger().Info("All servers have been shut down successfully.")

	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServe(parentCtx context.Context, wg *sync.WaitGroup) {
	notifLogger().Info(fmt.Sprintf("Starting gracefully shutdown for http server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()

	if err := app.Server.Stop(httpShutdownCtx); err != nil {
		notifLogger().Error(fmt.Sprintf("failed http server gracefully shutdown: %v", err))
	}

	notifLogger().Info("Successfully http server gracefully shutdown")
}

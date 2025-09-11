package leaderboardstatapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp/repository"
	"github.com/gocasters/rankr/pkg/database"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/httpserver"

	"github.com/gocasters/rankr/leaderboardstatapp/delivery/http"
	"github.com/gocasters/rankr/pkg/logger"
)

type Application struct {
	LeaderboardstatRepo    leaderboardstat.Repository
	LeaderboardstatSrv     leaderboardstat.Service
	LeaderboardstatHandler http.Handler
	HTTPServer             http.Server
	Config                 Config
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
) (Application, error) {

	leaderboardstatRepo := repository.NewLeaderboardstatRepo(config.Repository, postgresConn)
	leaderboardstatValidator := leaderboardstat.NewValidator(leaderboardstatRepo)
	leaderboardstatSvc := leaderboardstat.NewService(leaderboardstatRepo, leaderboardstatValidator)
	leaderboardstatHandler := http.NewHandler(leaderboardstatSvc)

	leaderboardLogger := logger.L()
	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		leaderboardLogger.Error("failed to initialize HTTP server", err)
		return Application{}, err
	}

	return Application{
		LeaderboardstatRepo:    leaderboardstatRepo,
		LeaderboardstatSrv:     leaderboardstatSvc,
		LeaderboardstatHandler: leaderboardstatHandler,
		HTTPServer: http.New(
			*httpServer,
			leaderboardstatHandler,
		),
		Config: config,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	<-ctx.Done()

	leaderboardLogger := logger.L()
	leaderboardLogger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		leaderboardLogger.Info("Servers shut down gracefully")
	} else {
		leaderboardLogger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	leaderboardLogger.Info("leaderboardstat_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	leaderboardLogger := logger.L()

	wg.Add(1)
	go func() {
		defer wg.Done()
		leaderboardLogger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			leaderboardLogger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		leaderboardLogger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	leaderboardLogger := logger.L()
	leaderboardLogger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		leaderboardLogger.Info("All servers have been shut down successfully.")

	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	leaderboardLogger := logger.L()
	leaderboardLogger.Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()

	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		leaderboardLogger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	leaderboardLogger.Info("HTTP server shut down successfully.")
}

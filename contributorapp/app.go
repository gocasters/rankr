package contributorapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/cachemanager"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/delivery/http"
	"github.com/gocasters/rankr/contributorapp/repository"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Application struct {
	ContributorRepo    contributor.Repository
	ContributorSrv     contributor.Service
	ContributorHandler http.Handler
	HTTPServer         http.Server
	Config             Config
	Logger             *slog.Logger
	Redis              *redis.Adapter
	CacheManager       cachemanager.CacheManager
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
	logger *slog.Logger,
) (Application, error) {

	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		logger.Error("failed to initialize Redis", "err", err)
		return Application{}, err
	}
	cache := cachemanager.NewCacheManager(redisAdapter)

	contributorRepo := repository.NewContributorRepo(config.Repository, postgresConn, logger)
	contributorValidator := contributor.NewValidator(contributorRepo)
	contributorSvc := contributor.NewService(contributorRepo, *cache, contributorValidator)

	contributorHandler := http.NewHandler(contributorSvc, logger)

	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		logger.Error("failed to initialize HTTP server", "err", err)
		return Application{}, err
	}
	return Application{
		ContributorRepo:    contributorRepo,
		ContributorSrv:     contributorSvc,
		ContributorHandler: contributorHandler,
		HTTPServer: http.New(
			*httpServer,
			contributorHandler,
			logger,
		),
		Config:       config,
		Logger:       logger,
		Redis:        redisAdapter,
		CacheManager: *cache,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	<-ctx.Done()
	app.Logger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		app.Logger.Info("Servers shut down gracefully")
	} else {
		app.Logger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	app.Logger.Info("contributor_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("✅ HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), err)
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	parentCtx := context.Background()
	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(parentCtx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		app.Logger.Info("All servers have been shut down successfully.")

	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()
	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		app.Logger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	app.Logger.Info("HTTP server shut down successfully.")

}

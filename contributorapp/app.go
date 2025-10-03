package contributorapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/cachemanager"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	Redis              *redis.Adapter
	CacheManager       cachemanager.CacheManager
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
) (Application, error) {

	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		logger.L().Error("failed to initialize Redis", "err", err)
		return Application{}, err
	}
	cache := cachemanager.NewCacheManager(redisAdapter)

	contributorRepo := repository.NewContributorRepo(config.Repository, postgresConn)
	contributorValidator := contributor.NewValidator(contributorRepo)
	contributorSvc := contributor.NewService(contributorRepo, *cache, contributorValidator)

	contributorHandler := http.NewHandler(contributorSvc)

	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		logger.L().Error("failed to initialize HTTP server", "err", err)
		return Application{}, err
	}
	return Application{
		ContributorRepo:    contributorRepo,
		ContributorSrv:     contributorSvc,
		ContributorHandler: contributorHandler,
		HTTPServer: http.New(
			*httpServer,
			contributorHandler,
		),
		Config:       config,
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
	logger.L().Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		logger.L().Info("Servers shut down gracefully")
	} else {
		logger.L().Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	logger.L().Info("contributor_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.L().Info(fmt.Sprintf("✅ HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			logger.L().Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), err)
		}
		logger.L().Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	logger.L().Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	parentCtx := context.Background()
	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(parentCtx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		logger.L().Info("All servers have been shut down successfully.")

	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	logger.L().Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()
	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		logger.L().Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	logger.L().Info("HTTP server shut down successfully.")
}

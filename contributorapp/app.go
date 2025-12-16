package contributorapp

import (
	"context"
	"fmt"
	middleware2 "github.com/gocasters/rankr/contributorapp/delivery/http/middleware"
	"github.com/gocasters/rankr/contributorapp/service/job"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/adapter/redis"
	contributorgrpc "github.com/gocasters/rankr/contributorapp/delivery/grpc"
	"github.com/gocasters/rankr/contributorapp/delivery/http"
	"github.com/gocasters/rankr/contributorapp/repository"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Application struct {
	ContributorRepo    contributor.Repository
	ContributorSrv     contributor.Service
	ContributorHandler http.Handler
	HTTPServer         http.Server
	GRPCServer         contributorgrpc.Server
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
	contributorValidator := contributor.NewValidator()
	contributorSvc := contributor.NewService(contributorRepo, *cache, contributorValidator)

	jobRepo := repository.NewJobRepository(postgresConn)
	failRecordRepo := repository.NewFailRecordRepository(postgresConn)
	broker := repository.NewBroker(config.Broker, redisAdapter)
	contributorAdapter := job.NewContributorAdapter(contributorSvc)
	validator := job.NewValidator(config.Validation)

	jobSvc := job.NewService(config.Job, jobRepo, broker, contributorAdapter, failRecordRepo, validator)

	contributorHandler := http.NewHandler(contributorSvc, jobSvc, logger)

	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		logger.Error("failed to initialize HTTP server", "err", err)
		return Application{}, err
	}

	middleware := middleware2.New(config.Middleware)

	grpcServer, err := grpc.NewServer(config.GRPCServer)
	if err != nil {
		logger.Error("failed to initialize gRPC server", "err", err)
		return Application{}, err
	}
	grpcHandler := contributorgrpc.NewHandler(contributorSvc)

	return Application{
		ContributorRepo:    contributorRepo,
		ContributorSrv:     contributorSvc,
		ContributorHandler: contributorHandler,
		HTTPServer: http.New(
			*httpServer,
			contributorHandler,
			logger,
			middleware,
		),
		GRPCServer:   contributorgrpc.New(grpcServer, grpcHandler),
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
	wg.Add(2)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), slog.String("error", err.Error()))
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()

	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("gRPC server started on %d", app.Config.GRPCServer.Port))
		if err := app.GRPCServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("error in gRPC server on %d", app.Config.GRPCServer.Port), slog.String("error", err.Error()))
		}
		app.Logger.Info(fmt.Sprintf("gRPC server stopped %d", app.Config.GRPCServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	parentCtx := context.Background()
	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(2)
		go app.shutdownHTTPServer(parentCtx, &shutdownWg)
		go app.shutdownGRPCServer(&shutdownWg)

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

func (app Application) shutdownGRPCServer(wg *sync.WaitGroup) {
	app.Logger.Info(fmt.Sprintf("Starting graceful shutdown for gRPC server on port %d", app.Config.GRPCServer.Port))

	defer wg.Done()
	app.GRPCServer.Stop()

	app.Logger.Info("gRPC server shut down successfully.")
}

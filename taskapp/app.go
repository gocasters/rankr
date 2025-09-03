package taskapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/cachemanager"
	"github.com/gocasters/rankr/taskapp/service/task"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/taskapp/delivery/http"
	"github.com/gocasters/rankr/taskapp/repository"
)

type Application struct {
	TaskRepo     task.Repository
	TaskSrv      task.Service
	TaskHandler  http.Handler
	HTTPServer   http.Server
	Config       Config
	Logger       *slog.Logger
	Redis        *redis.Adapter
	CacheManager cachemanager.CacheManager
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
	logger *slog.Logger,
) Application {

	redisAdapter, _ := redis.New(ctx, config.Redis)
	cache := cachemanager.NewCacheManager(redisAdapter)

	taskRepo := repository.NewTaskRepo(config.Repository, postgresConn, logger)
	taskValidator := task.NewValidator(taskRepo)
	taskSvc := task.NewService(taskRepo, *cache, taskValidator, logger)

	taskHandler := http.NewHandler(taskSvc, logger)

	httpServer, _ := httpserver.New(config.HTTPServer)

	return Application{
		TaskRepo:    taskRepo,
		TaskSrv:     taskSvc,
		TaskHandler: taskHandler,
		HTTPServer: http.New(
			*httpServer,
			taskHandler,
			logger,
		),
		Config: config,
		Logger: logger,
	}
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
	app.Logger.Info("task_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("✅ HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	parentCtx := ctx
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

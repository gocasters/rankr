package projectapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/projectapp/delivery/http"
	"github.com/gocasters/rankr/projectapp/repository"
	"github.com/gocasters/rankr/projectapp/service/project"
	"github.com/gocasters/rankr/projectapp/service/versioncontrollersystemproject"
)

type Application struct {
	ProjectRepo                        project.Repository
	versionControllerSystemProjectRepo versioncontrollersystemproject.Repository

	ProjectService                        project.Service
	VersionControllerSystemProjectService versioncontrollersystemproject.Service

	HTTPServer http.Server

	Config Config

	Logger *slog.Logger
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
	logger *slog.Logger,
) Application {
	if postgresConn == nil || postgresConn.Pool == nil {
		logger.Error("Postgres connection pool is not initialized")
		panic("postgres connection pool is nil")
	}

	projectRepo := repository.NewProjectRepository(postgresConn)

	projectValidator := project.NewValidator(projectRepo)

	projectService := project.NewService(projectRepo, projectValidator, logger)

	httpServer, hErr := httpserver.New(config.HTTPServer)
	if hErr != nil {
		logger.Error("Failed to initialize HTTP server", slog.String("error", hErr.Error()))
		panic(hErr)
	}

	versionControllerSystemProjectRepo := repository.NewVersionControllerSystemProjectRepository(postgresConn)
	versionSystemProjectValidator := versioncontrollersystemproject.NewValidator()
	versionSystemProjectService := versioncontrollersystemproject.NewService(versionControllerSystemProjectRepo, versionSystemProjectValidator, logger)

	projectHandler := http.NewHandler(projectService, versionSystemProjectService, logger)
	projectHttpService := http.New(
		httpServer,
		projectHandler,
		logger,
		projectService,
		versionSystemProjectService,
	)

	return Application{
		ProjectRepo:                           projectRepo,
		versionControllerSystemProjectRepo:    versionControllerSystemProjectRepo,
		ProjectService:                        projectService,
		VersionControllerSystemProjectService: versionSystemProjectService,
		HTTPServer:                            projectHttpService,
		Config:                                config,
		Logger:                                logger,
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
	app.Logger.Info("projects_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error("error in serving HTTP server projects app", "error", err)
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

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

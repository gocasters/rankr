package authapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	statHTTP "github.com/gocasters/rankr/authapp/delivery/http"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/authapp/repository"
	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
)

type Application struct {
	Repo       auth.Repository
	Srv        auth.Service
	Handler    statHTTP.Handler
	HTTPServer statHTTP.Server
	Config     Config
	Validator  auth.Validator
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
) (Application, error) {
	log := logger.L()

	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		log.Error("failed to initialize Redis", "err", err)
		return Application{}, err
	}
	cache := cachemanager.NewCacheManager(redisAdapter)

	repo := repository.NewAuthRepo(config.Repository, postgresConn)
	validator := auth.NewValidator(repo)
	svc := auth.NewService(repo, validator, *cache, nil)

	httpSrvCore, err := httpserver.New(config.HTTPServer)
	if err != nil {
		log.Error("failed to initialize HTTP server", slog.Any("error", err))
		return Application{}, err
	}

	httpHandler := statHTTP.NewHandler(svc)
	httpSrv := statHTTP.New(*httpSrvCore, httpHandler)

	return Application{
		Repo:       repo,
		Srv:        svc,
		Handler:    httpHandler,
		HTTPServer: httpSrv,
		Config:     config,
		Validator:  validator,
	}, nil
}

func (app Application) Start() {
	log := logger.L()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	startServers(app, &wg)

	<-ctx.Done()
	log.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		log.Info("Servers shut down gracefully")
	} else {
		log.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	log.Info("auth_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	log := logger.L()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info(fmt.Sprintf("HTTP Server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			log.Error(fmt.Sprintf("error in authapp HTTP Server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		log.Info(fmt.Sprintf("HTTP Server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	log := logger.L()
	log.Info("Starting authapp server shutdown process...")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		log.Info("HTTP server has been shut down successfully.")
	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log := logger.L()
	log.Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))

	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()

	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		log.Error("HTTP server graceful shutdown failed", slog.Any("error", err))
		return
	}

	log.Info("HTTP server shut down successfully.")
}

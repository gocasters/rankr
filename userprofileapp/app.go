package userprofileapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/userprofileapp/adapter"
	"github.com/gocasters/rankr/userprofileapp/delivery/http"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var userProfileLogger = logger.L()

type Application struct {
	Validator  userprofile.Validator
	RPCAdapter adapter.RPCAdapter
	Service    userprofile.Service
	Handler    *http.Handler
	HTTPServer http.Server
	Config     Config
}

func Setup(cfg Config) (Application, error) {
	rpcAdapter := adapter.NewRPCAdapter()
	validator := userprofile.NewValidator(rpcAdapter)
	service := userprofile.NewService(rpcAdapter, validator)
	handler := http.NewHandler(service)

	HTTPServer, err := httpserver.New(cfg.HTTPServer)
	if err != nil {
		logger.L().Error("failed to initialize http server", "error", err)
		return Application{}, err
	}

	httpServer := http.New(HTTPServer, handler)

	return Application{
		Validator:  validator,
		RPCAdapter: rpcAdapter,
		Service:    service,
		Handler:    handler,
		HTTPServer: httpServer,
		Config:     cfg,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	<-ctx.Done()

	userProfileLogger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		userProfileLogger.Info("Servers shutdown gracefully")
	} else {
		userProfileLogger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	userProfileLogger.Info("user profile app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		userProfileLogger.Info(fmt.Sprintf("HTTP server starting on port %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			userProfileLogger.Error(fmt.Sprintf("error listen and serve http server on port %d",
				&app.Config.HTTPServer.Port), "error", err)
		}

		userProfileLogger.Info(fmt.Sprintf("Http server stopped on port %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	userProfileLogger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServe(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		userProfileLogger.Info("All servers have been shut down successfully.")

	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServe(parentCtx context.Context, wg *sync.WaitGroup) {
	userProfileLogger.Info(fmt.Sprintf("Starting gracefully shutdown for http server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()

	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		userProfileLogger.Error(fmt.Sprintf("failed http server gracefully shutdown: %v", err))
	}

	userProfileLogger.Info("Successfully http server gracefully shutdown")
}

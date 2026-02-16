package authapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/adapter/contributor"
	statHTTP "github.com/gocasters/rankr/authapp/delivery/http"
	"github.com/gocasters/rankr/authapp/repository"
	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/authapp/service/tokenservice"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
)

type Application struct {
	ShutdownCtx  context.Context
	Repo         auth.Repository
	Srv          auth.Service
	TokenService *tokenservice.AuthService
	Contributor  *contributor.Client
	Handler      statHTTP.Handler
	HTTPServer   statHTTP.Server
	Config       Config
	Validator    auth.Validator
}

type contributorCredentialsProvider interface {
	VerifyPassword(ctx context.Context, username string, password string) (types.ID, string, bool, error)
}

type unavailableContributorClient struct{}

func (unavailableContributorClient) VerifyPassword(_ context.Context, _, _ string) (types.ID, string, bool, error) {
	return 0, "", false, fmt.Errorf("contributor service is unavailable")
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
) (Application, error) {
	log := logger.L()

	repo := repository.NewRepository(postgresConn)
	validator := auth.NewValidator(repo)

	tokenSvc := tokenservice.NewAuthService(config.JWT.Secret, config.JWT.TokenDuration, config.JWT.RefreshTokenDuration)

	var contributorClient *contributor.Client
	var contributorProvider contributorCredentialsProvider = unavailableContributorClient{}

	rpcClient, rpcErr := grpc.NewClient(config.ContributorRPC, log)
	if rpcErr != nil {
		log.Warn("failed to initialize contributor RPC client; auth will start in degraded mode", slog.Any("error", rpcErr))
	} else {
		contributorClient, contributorErr := contributor.New(rpcClient)
		if contributorErr != nil {
			rpcClient.Close()
			log.Warn("failed to initialize contributor client; auth will start in degraded mode", slog.Any("error", contributorErr))
		} else {
			contributorProvider = contributorClient
		}
	}
	svc := auth.NewService(repo, validator, contributorProvider, tokenSvc)
	httpSrvCore, err := httpserver.New(config.HTTPServer)
	if err != nil {
		log.Error("failed to initialize HTTP server", slog.Any("error", err))
		return Application{}, err
	}
	httpHandler := statHTTP.NewHandler(svc, tokenSvc)
	httpSrv := statHTTP.New(*httpSrvCore, httpHandler)

	return Application{
		ShutdownCtx:  ctx,
		Repo:         repo,
		Srv:          svc,
		TokenService: tokenSvc,
		Contributor:  contributorClient,
		Handler:      httpHandler,
		HTTPServer:   httpSrv,
		Config:       config,
		Validator:    validator,
	}, nil
}

func (app Application) Start() {
	log := logger.L()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	startServers(app, &wg)

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		log.Info("servers shut down gracefully")
	} else {
		log.Warn("shutdown timed out; forcing exit")
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
		log.Info("HTTP server starting", slog.Int("port", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			log.Error("HTTP server error", slog.Int("port", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		log.Info("HTTP server stopped", slog.Int("port", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	log := logger.L()
	log.Info("starting authapp server shutdown process")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		log.Info("HTTP server has been shut down successfully")
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
	log.Info("starting graceful shutdown for HTTP server", slog.Int("port", app.Config.HTTPServer.Port))

	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()

	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		log.Error("HTTP server graceful shutdown failed", slog.Any("error", err))
		return
	}

	log.Info("HTTP server shut down successfully")

	if app.Contributor != nil {
		app.Contributor.Close()
	}
}

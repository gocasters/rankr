package leaderboardstatapp

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/leaderboardscoring"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardstatapp/delivery/scheduler"
	"github.com/gocasters/rankr/leaderboardstatapp/repository"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/database"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/httpserver"

	statGRPC "github.com/gocasters/rankr/leaderboardstatapp/delivery/grpc"
	statHTTP "github.com/gocasters/rankr/leaderboardstatapp/delivery/http"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
)

type Application struct {
	LeaderboardstatRepo    leaderboardstat.Repository
	LeaderboardstatSrv     leaderboardstat.Service
	LeaderboardstatHandler statHTTP.Handler
	HTTPServer             statHTTP.Server
	GRPCServer             statGRPC.Server
	Config                 Config
	CacheManager           cachemanager.CacheManager
	redis                  *redis.Adapter
	Scheduler              scheduler.Scheduler
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
) (Application, error) {
	statLogger := logger.L()

	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		statLogger.Error("failed to initialize Redis", "err", err)
		return Application{}, err
	}
	cache := cachemanager.NewCacheManager(redisAdapter)

	// initial rpc server
	rpcClient, err := grpc.NewClient(config.LeaderboardScoringRPC, statLogger)
	if err != nil {
		return Application{}, fmt.Errorf("failed to create RPC client!!: %w", err)
	}

	lbScoringClient, err := leaderboardscoring.New(rpcClient)
	if err != nil {
		return Application{}, fmt.Errorf("failed to create leaderboardscoring client: %w", err)
	}

	statRepo := repository.NewLeaderboardstatRepo(config.Repository, postgresConn)
	statValidator := leaderboardstat.NewValidator(statRepo)

	statSvc := leaderboardstat.NewService(statRepo, statValidator, *cache, nil, lbScoringClient)
	statHandler := statHTTP.NewHandler(statSvc)

	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		statLogger.Error("failed to initialize HTTP server", slog.Any("error", err))
		return Application{}, err
	}

	rpcServer, gErr := grpc.NewServer(config.RPCServer)
	if gErr != nil {
		statLogger.Error("Failed to initialize gRPC server", slog.String("error", gErr.Error()))
		return Application{}, gErr
	}
	statGrpcHandler := statGRPC.NewHandler(statSvc)
	statGrpcServer := statGRPC.New(rpcServer, statGrpcHandler)

	// Initialize scheduler
	statScheduler := scheduler.New(&statSvc, config.SchedulerCfg)

	return Application{
		LeaderboardstatRepo:    statRepo,
		LeaderboardstatSrv:     statSvc,
		LeaderboardstatHandler: statHandler,
		HTTPServer: statHTTP.New(
			*httpServer,
			statHandler,
		),
		GRPCServer:   statGrpcServer,
		Config:       config,
		CacheManager: *cache,
		redis:        redisAdapter,
		Scheduler:    statScheduler,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	app.startScheduler(ctx, &wg)
	<-ctx.Done()

	statLogger := logger.L()
	statLogger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		statLogger.Info("Servers shut down gracefully")
	} else {
		statLogger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	statLogger.Info("leaderboardstat_app stopped")
}

func (app Application) startScheduler(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		app.Scheduler.Start(ctx, wg)
	}()
}

func startServers(app Application, wg *sync.WaitGroup) {
	statLogger := logger.L()

	wg.Add(1)
	go func() {
		defer wg.Done()
		statLogger.Info(fmt.Sprintf("HTTP Server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			statLogger.Error(fmt.Sprintf("error in leaderboard-stat HTTP Server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		statLogger.Info(fmt.Sprintf("HTTP Server stopped %d", app.Config.HTTPServer.Port))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("grpc server started")
		statLogger.Info("gRPC Server started")
		if err := app.GRPCServer.Serve(); err != nil {
			statLogger.Error("error in serving leaderboard-stat gRPC server", "error", err)
		}
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	statLogger := logger.L()
	statLogger.Info("Starting leaderboard-stat server shutdown process...")
	fmt.Println("grpc server stoppped...")
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(2)
		go app.shutdownHTTPServer(ctx, &shutdownWg)
		go app.shutdownGRPCServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		statLogger.Info("All servers have been shut down successfully.")

	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	statLogger := logger.L()
	statLogger.Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()

	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		statLogger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	statLogger.Info("HTTP server shut down successfully.")
}

func (app Application) shutdownGRPCServer(parentCtx context.Context, wg *sync.WaitGroup) {
	statLogger := logger.L()
	defer wg.Done()
	statLogger.Info("starting gracefully shutdown leaderboard-stat gRPC server")

	app.GRPCServer.Stop()

	statLogger.Info("leaderboard-stat gRPC server shutdown successfully.")
}

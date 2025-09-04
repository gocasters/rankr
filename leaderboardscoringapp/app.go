package leaderboardscoringapp

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer"
	leaderboardGRPC "github.com/gocasters/rankr/leaderboardscoringapp/delivery/grpc"
	leaderboardHTTP "github.com/gocasters/rankr/leaderboardscoringapp/delivery/http"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/scheduler"
	"github.com/gocasters/rankr/leaderboardscoringapp/repository"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type Application struct {
	HTTPServer                leaderboardHTTP.Server
	LeaderboardGrpcServer     leaderboardGRPC.Server
	Scheduler                 *scheduler.Scheduler
	LeaderboardSvc            leaderboardscoring.Service
	LeaderboardscoringHandler leaderboardHTTP.Handler
	WMRouter                  *message.Router
	Logger                    *slog.Logger
	WMLogger                  watermill.LoggerAdapter
	Config                    Config
	Subscriber                message.Subscriber
	RedisAdapter              *redis.Adapter
}

func Setup(ctx context.Context, config Config, leaderboardLogger *slog.Logger,
	subscriber message.Subscriber, postgresConn *database.Database, wmLogger watermill.LoggerAdapter) *Application {

	if strings.TrimSpace(config.SubscriberTopic) == "" {
		leaderboardLogger.Error("SubscriberTopic is empty; set config.subscriber_topic")
		panic("invalid config: subscriber_topic")
	}

	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		leaderboardLogger.Error("Failed to initialize Redis adapter", slog.String("error", err.Error()))
		panic(err)
	}

	if postgresConn == nil || postgresConn.Pool == nil {
		leaderboardLogger.Error("Postgres connection pool is not initialized")
		panic("postgres connection pool is nil")
	}

	lbScoringRepo := repository.NewLeaderboardscoringRepo(redisAdapter.Client(), postgresConn.Pool, leaderboardLogger)
	lbScoringValidator := leaderboardscoring.NewValidator()
	lbScoringService := leaderboardscoring.NewService(lbScoringRepo, lbScoringValidator, leaderboardLogger, config.LeaderboardScoring)

	httpServer, hErr := httpserver.New(config.HTTPServer)
	if hErr != nil {
		leaderboardLogger.Error("Failed to initialize HTTP server", slog.String("error", hErr.Error()))
		panic(hErr)
	}
	lbScoringHandler := leaderboardHTTP.NewHandler(leaderboardLogger)
	leaderboardHttpServer := leaderboardHTTP.New(
		httpServer,
		lbScoringHandler,
		leaderboardLogger,
		lbScoringService,
	)

	sch := scheduler.New(lbScoringService, leaderboardLogger, config.Scheduler)

	rpcServer, gErr := grpc.NewServer(config.RPCServer, leaderboardLogger)
	if gErr != nil {
		leaderboardLogger.Error("Failed to initialize gRPC server", slog.String("error", gErr.Error()))
		panic(gErr)
	}
	leaderboardGrpcHandler := leaderboardGRPC.NewHandler(lbScoringService, leaderboardLogger)
	leaderboardGrpcServer := leaderboardGRPC.New(rpcServer, leaderboardGrpcHandler, leaderboardLogger)

	return &Application{
		HTTPServer:                leaderboardHttpServer,
		LeaderboardGrpcServer:     leaderboardGrpcServer,
		Scheduler:                 sch,
		LeaderboardSvc:            lbScoringService,
		LeaderboardscoringHandler: lbScoringHandler,
		WMRouter:                  nil,
		Logger:                    leaderboardLogger,
		WMLogger:                  wmLogger,
		Config:                    config,
		Subscriber:                subscriber,
		RedisAdapter:              redisAdapter,
	}
}

func (app *Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.startHTTPServer(&wg)
	app.startWaterMill(ctx, &wg)
	app.startScheduler(ctx, &wg)
	app.startGRPCServer(&wg)

	app.Logger.Info("Leaderboard Scoring application started.")

	<-ctx.Done()
	app.Logger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		app.Logger.Info("Servers shutdown gracefully")
	} else {
		app.Logger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	app.Logger.Info("leaderboard-scoring stopped")
}

func (app *Application) startHTTPServer(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("HTTP server starting...",
			slog.String("host", app.Config.HTTPServer.Host),
			slog.Int("port", app.Config.HTTPServer.Port))

		if err := app.HTTPServer.Serve(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				app.Logger.Error(
					"HTTP server failed",
					slog.Int("port", app.Config.HTTPServer.Port),
					slog.Any("error", err),
				)
				panic(err)
			}
		}

		app.Logger.Info("HTTP server stopped",
			slog.String("host", app.Config.HTTPServer.Host),
			slog.Int("port", app.Config.HTTPServer.Port))
	}()
}

func (app *Application) startWaterMill(ctx context.Context, wg *sync.WaitGroup) {
	app.setupWatermill()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Starting Watermill event consumer router...")

		if err := app.WMRouter.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.Logger.Error("Watermill router stopped with an error", slog.String("error", err.Error()))
			panic(err)

		}
	}()
}

func (app *Application) setupWatermill() {

	router, err := message.NewRouter(message.RouterConfig{}, app.WMLogger)
	if err != nil {
		app.Logger.Error("Failed to create Watermill router", slog.String("error", err.Error()))
		panic(err)
	}

	checker := consumer.NewIdempotencyChecker(app.RedisAdapter.Client(), app.Config.Consumer, app.Logger)
	handler := consumer.NewHandler(app.LeaderboardSvc, checker, app.Logger)

	router.AddNoPublisherHandler(
		"ContributionHandler",
		app.Config.SubscriberTopic,
		app.Subscriber,
		handler.HandleEvent,
	)

	app.WMRouter = router
}

func (app *Application) startScheduler(ctx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info("Starting Scheduler...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Scheduler.Start(ctx)
	}()
}

func (app *Application) startGRPCServer(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.LeaderboardGrpcServer.Serve(); err != nil {
			app.Logger.Error("error in serving leaderboard-scoring gRPC server", "error", err)
		}
	}()
}

func (app *Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup

		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Add(1)
		go app.shutdownWatermillAndSubscriber(ctx, &shutdownWg)

		app.Scheduler.Stop()

		shutdownWg.Add(1)
		go app.shutdownGRPCServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		app.Logger.Info("All servers have been shutdown successfully.")
	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app *Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	app.Logger.Info("Starting graceful shutdown for HTTP server", "port", app.Config.HTTPServer.Port)

	httpCtx, cancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := app.HTTPServer.HTTPServer.Stop(httpCtx); err != nil {
		app.Logger.Error("HTTP server graceful shutdown failed", "error", err)
	} else {
		app.Logger.Info("HTTP server shutdown successfully")
	}
}

func (app *Application) shutdownWatermillAndSubscriber(parentCtx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	app.Logger.Info("Starting graceful shutdown for Watermill router and subscriber...")

	done := make(chan struct{})
	go func() {
		// first shutdown gracefully watermill router
		app.Logger.Info("Closing Watermill router...")
		if err := app.WMRouter.Close(); err != nil {
			app.Logger.Error("Watermill router graceful shutdown failed", "error", err)
		} else {
			app.Logger.Info("Watermill router shutdown successfully.")
		}

		// second close subscriber
		app.Logger.Info("Closing subscriber...")
		if err := app.Subscriber.Close(); err != nil {
			app.Logger.Error("Subscriber close failed", "error", err)
		} else {
			app.Logger.Info("Subscriber closed successfully.")
		}

		close(done)
	}()

	select {
	case <-done:
		app.Logger.Info("Watermill and subscriber shutdown completed.")
	case <-parentCtx.Done():
		app.Logger.Warn("Watermill and subscriber shutdown timed out.")
	}
}

func (app *Application) shutdownGRPCServer(parentCtx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	app.Logger.Info("starting gracefully shutdown leaderboard-scoring gRPC server")

	app.LeaderboardGrpcServer.Stop()

	app.Logger.Info("leaderboard-scoring gRPC server shutdown successfully.")
}

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
	"github.com/gocasters/rankr/leaderboardscoringapp/repository"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
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
	LeaderboardSvc            leaderboardscoring.Service
	LeaderboardscoringHandler leaderboardHTTP.Handler
	WMRouter                  *message.Router
	WMLogger                  watermill.LoggerAdapter
	Config                    Config
	Subscriber                message.Subscriber
	RedisAdapter              *redis.Adapter
}

func Setup(ctx context.Context, config Config, subscriber message.Subscriber,
	postgresConn *database.Database, wmLogger watermill.LoggerAdapter) *Application {
	logger := logger.L()

	if subscriber == nil {
		logger.Error("subscriber is nil; provide a valid message.Subscriber")
		panic("invalid config: subscriber is nil")
	}

	if wmLogger == nil {
		logger.Error("watermill logger is nil; provide a valid LoggerAdapter")
		panic("invalid config: watermill logger is nil")
	}

	if strings.TrimSpace(config.SubscriberTopic) == "" {
		logger.Error("SubscriberTopic is empty; set config.subscriber_topic")
		panic("invalid config: subscriber_topic")
	}

	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		logger.Error("Failed to initialize Redis adapter", slog.String("error", err.Error()))
		panic(err)
	}

	if postgresConn == nil || postgresConn.Pool == nil {
		logger.Error("Postgres connection pool is not initialized")
		panic("postgres connection pool is nil")
	}

	lbScoringRepo := repository.NewLeaderboardscoringRepo(redisAdapter.Client(), postgresConn.Pool)
	lbScoringValidator := leaderboardscoring.NewValidator()
	lbScoringService := leaderboardscoring.NewService(lbScoringRepo, lbScoringValidator)

	httpServer, hErr := httpserver.New(config.HTTPServer)
	if hErr != nil {
		logger.Error("Failed to initialize HTTP server", slog.String("error", hErr.Error()))
		panic(hErr)
	}
	lbScoringHandler := leaderboardHTTP.NewHandler()
	leaderboardHttpServer := leaderboardHTTP.New(httpServer, lbScoringHandler, lbScoringService)

	rpcServer, gErr := grpc.NewServer(config.RPCServer)
	if gErr != nil {
		logger.Error("Failed to initialize gRPC server", slog.String("error", gErr.Error()))
		panic(gErr)
	}
	leaderboardGrpcHandler := leaderboardGRPC.NewHandler(lbScoringService)
	leaderboardGrpcServer := leaderboardGRPC.New(rpcServer, leaderboardGrpcHandler)

	return &Application{
		HTTPServer:                leaderboardHttpServer,
		LeaderboardGrpcServer:     leaderboardGrpcServer,
		LeaderboardSvc:            lbScoringService,
		LeaderboardscoringHandler: lbScoringHandler,
		WMRouter:                  nil,
		WMLogger:                  wmLogger,
		Config:                    config,
		Subscriber:                subscriber,
		RedisAdapter:              redisAdapter,
	}
}

func (app *Application) Start() {
	logger := logger.L()
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.startHTTPServer(&wg)
	app.startWaterMill(ctx, &wg)
	app.startGRPCServer(&wg)

	logger.Info("Leaderboard Scoring application started.")

	<-ctx.Done()
	logger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		logger.Info("Servers shutdown gracefully")
	} else {
		logger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	logger.Info("leaderboard-scoring stopped")
}

func (app *Application) startHTTPServer(wg *sync.WaitGroup) {
	logger := logger.L()
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("HTTP server starting...",
			slog.String("host", app.Config.HTTPServer.Host),
			slog.Int("port", app.Config.HTTPServer.Port))

		if err := app.HTTPServer.Serve(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error(
					"HTTP server failed",
					slog.Int("port", app.Config.HTTPServer.Port),
					slog.Any("error", err),
				)
				panic(err)
			}
		}

		logger.Info("HTTP server stopped",
			slog.String("host", app.Config.HTTPServer.Host),
			slog.Int("port", app.Config.HTTPServer.Port))
	}()
}

func (app *Application) startWaterMill(ctx context.Context, wg *sync.WaitGroup) {
	logger := logger.L()
	app.setupWatermill()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting Watermill event consumer router...")

		if err := app.WMRouter.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Watermill router stopped with an error", slog.String("error", err.Error()))
			panic(err)

		}
	}()
}

func (app *Application) setupWatermill() {
	logger := logger.L()
	router, err := message.NewRouter(message.RouterConfig{}, app.WMLogger)
	if err != nil {
		logger.Error("Failed to create Watermill router", slog.String("error", err.Error()))
		panic(err)
	}

	checker := consumer.NewIdempotencyChecker(app.RedisAdapter.Client(), app.Config.Consumer)
	handler := consumer.NewHandler(app.LeaderboardSvc, checker)

	router.AddNoPublisherHandler(
		"ContributionHandler",
		app.Config.SubscriberTopic,
		app.Subscriber,
		handler.HandleEvent,
	)

	app.WMRouter = router
}

func (app *Application) startGRPCServer(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.LeaderboardGrpcServer.Serve(); err != nil {
			logger.L().Error("error in serving leaderboard-scoring gRPC server", "error", err)
		}
	}()
}

func (app *Application) shutdownServers(ctx context.Context) bool {
	logger := logger.L()
	logger.Info("Starting server shutdown process...")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup

		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Add(1)
		go app.shutdownWatermillAndSubscriber(ctx, &shutdownWg)

		shutdownWg.Add(1)
		go app.shutdownGRPCServer(ctx, &shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		logger.Info("All servers have been shutdown successfully.")
	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app *Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	logger := logger.L()
	defer wg.Done()
	logger.Info("Starting graceful shutdown for HTTP server", "port", app.Config.HTTPServer.Port)

	httpCtx, cancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := app.HTTPServer.HTTPServer.Stop(httpCtx); err != nil {
		logger.Error("HTTP server graceful shutdown failed", "error", err)
	} else {
		logger.Info("HTTP server shutdown successfully")
	}
}

func (app *Application) shutdownWatermillAndSubscriber(parentCtx context.Context, wg *sync.WaitGroup) {
	logger := logger.L()
	defer wg.Done()

	logger.Info("Starting graceful shutdown for Watermill router and subscriber...")

	done := make(chan struct{})
	go func() {
		// first shutdown gracefully watermill router
		logger.Info("Closing Watermill router...")
		if err := app.WMRouter.Close(); err != nil {
			logger.Error("Watermill router graceful shutdown failed", "error", err)
		} else {
			logger.Info("Watermill router shutdown successfully.")
		}

		// second close subscriber
		logger.Info("Closing subscriber...")
		if err := app.Subscriber.Close(); err != nil {
			logger.Error("Subscriber close failed", "error", err)
		} else {
			logger.Info("Subscriber closed successfully.")
		}

		close(done)
	}()

	select {
	case <-done:
		logger.Info("Watermill and subscriber shutdown completed.")
	case <-parentCtx.Done():
		logger.Warn("Watermill and subscriber shutdown timed out.")
	}
}

func (app *Application) shutdownGRPCServer(parentCtx context.Context, wg *sync.WaitGroup) {
	logger := logger.L()
	defer wg.Done()
	logger.Info("starting gracefully shutdown leaderboard-scoring gRPC server")

	app.LeaderboardGrpcServer.Stop()

	logger.Info("leaderboard-scoring gRPC server shutdown successfully.")
}

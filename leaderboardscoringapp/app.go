package leaderboardscoringapp

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/http"
	"github.com/gocasters/rankr/leaderboardscoringapp/repository"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
)

type Application struct {
	HTTPServer                http.Server
	LeaderboardSvc            leaderboardscoring.Service
	LeaderboardscoringHandler http.Handler
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

	leaderboardscoringRepo := repository.NewLeaderboardscoringRepo(redisAdapter.Client(), postgresConn.Pool, leaderboardLogger)
	leaderboardValidator := leaderboardscoring.NewValidator()
	leaderboardSvc := leaderboardscoring.NewService(leaderboardscoringRepo, leaderboardValidator, leaderboardLogger)

	httpServer, hErr := httpserver.New(config.HTTPServer)
	if hErr != nil {
		leaderboardLogger.Error("Failed to initialize HTTP server", slog.String("error", hErr.Error()))
		panic(hErr)
	}
	leaderboardscoringHandler := http.NewHandler(leaderboardLogger)
	leaderboardHttpServer := http.New(
		httpServer,
		leaderboardscoringHandler,
		leaderboardLogger,
		leaderboardSvc,
	)

	return &Application{
		HTTPServer:                leaderboardHttpServer,
		LeaderboardSvc:            leaderboardSvc,
		LeaderboardscoringHandler: leaderboardscoringHandler,
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
	app.startWaterMill(&wg, ctx)

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
			// todo add metrics
			app.Logger.Error(
				fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port),
				slog.Any("error", err))

			panic(err)
		}

		app.Logger.Info("HTTP server stopped",
			slog.String("host", app.Config.HTTPServer.Host),
			slog.Int("port", app.Config.HTTPServer.Port))
	}()
}

func (app *Application) startWaterMill(wg *sync.WaitGroup, ctx context.Context) {
	app.setupWatermill()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Starting Watermill event consumer router...")

		if err := app.WMRouter.Run(ctx); err != nil {
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
		handler.HandleContributionRegistered,
	)

	app.WMRouter = router
}

func (app *Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Add(1)
		go app.shutdownWatermill(ctx, &shutdownWg)

		shutdownWg.Add(1)
		go app.closeSubscriber(ctx, &shutdownWg)

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
	app.Logger.Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on %s:%d", app.Config.HTTPServer.Host, app.Config.HTTPServer.Port))

	defer wg.Done()
	httpCtx, cancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := app.HTTPServer.HTTPServer.Stop(httpCtx); err != nil {
		app.Logger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	app.Logger.Info("HTTP server shutdown successfully")
}

func (app *Application) shutdownWatermill(parentCtx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info("Starting graceful shutdown for Watermill")
	defer wg.Done()

	done := make(chan struct{})
	go func() {
		if err := app.WMRouter.Close(); err != nil {
			app.Logger.Error("Watermill graceful shutdown failed")
		}
		close(done)
	}()

	select {
	case <-done:
		app.Logger.Info("Watermill shutdown successfully.")
	case <-parentCtx.Done():
		app.Logger.Warn("Watermill shutdown timed out.")
	}
}

func (app *Application) closeSubscriber(parentCtx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info("Close subscriber")
	defer wg.Done()
	done := make(chan struct{})
	go func() {
		if err := app.Subscriber.Close(); err != nil {
			app.Logger.Error("Close subscriber failed")
		}
		close(done)
	}()

	select {
	case <-done:
		app.Logger.Info("Close subscriber successfully.")
	case <-parentCtx.Done():
		app.Logger.Warn("Close subscriber timed out.")
	}
}

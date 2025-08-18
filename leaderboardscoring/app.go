package leaderboardscoring

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/leaderboardscoring/delivery/http"
	"github.com/gocasters/rankr/leaderboardscoring/repository"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoring/delivery/consumer"
	"github.com/gocasters/rankr/leaderboardscoring/service/leaderboardscoring"
)

type Application struct {
	HTTPServer                http.Server
	LeaderboardscoringHandler http.Handler
	WatermillRouter           *message.Router
	Logger                    *slog.Logger
	Config                    Config
}

func Setup(
	ctx context.Context,
	config Config,
	logger *slog.Logger,
	subscriber message.Subscriber,
	postgresConn *database.Database,
) Application {
	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		logger.Error("Failed to initialize Redis adapter", slog.String("error", err.Error()))
		panic(err)
	}

	if postgresConn == nil || postgresConn.Pool == nil {
		logger.Error("Postgres connection pool is not initialized")
		panic("postgres connection pool is nil")
	}

	leaderboardscoringRepo := repository.NewLeaderboardscoringRepo(redisAdapter.Client(), postgresConn.Pool, logger)
	leaderboardValidator := leaderboardscoring.NewValidator()
	leaderboardSvc := leaderboardscoring.NewService(leaderboardscoringRepo, leaderboardValidator, logger)

	httpServer, hErr := httpserver.New(config.HTTPServer)
	if hErr != nil {
		logger.Error("Failed to initialize HTTP server", slog.String("error", hErr.Error()))
		panic(hErr)
	}
	leaderboardscoringHandler := http.NewHandler(logger)
	leaderboardHttpServer := http.New(
		httpServer,
		leaderboardscoringHandler,
		logger,
		leaderboardSvc,
	)

	wmRouter := setupWatermill(leaderboardSvc, subscriber, config.Consumer, logger, redisAdapter)

	return Application{
		HTTPServer:                leaderboardHttpServer,
		LeaderboardscoringHandler: leaderboardscoringHandler,
		WatermillRouter:           wmRouter,
		Logger:                    logger,
		Config:                    config,
	}
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServer(app, &wg)
	startWaterMill(app, &wg, ctx)

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

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")

	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(ctx, &shutdownWg)

		shutdownWg.Add(1)
		go app.shutdownWatermill(ctx, &shutdownWg)

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

func (app Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info(fmt.Sprintf("Starting graceful shutdown for HTTP server on port %d", app.Config.HTTPServer.Port))

	defer wg.Done()
	httpCtx, cancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := app.HTTPServer.HTTPServer.Stop(httpCtx); err != nil {
		app.Logger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	app.Logger.Info("HTTP server shut down successfully.")
}

func (app Application) shutdownWatermill(parentCtx context.Context, wg *sync.WaitGroup) {
	app.Logger.Info("Starting graceful shutdown for Watermill")
	defer wg.Done()

	done := make(chan struct{})
	go func() {
		if err := app.WatermillRouter.Close(); err != nil {
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

func startServer(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			// todo add metrics
			app.Logger.Error(
				fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port),
				slog.Any("error", err))
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func startWaterMill(app Application, wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Starting Watermill event consumer router...")
		if err := app.WatermillRouter.Run(ctx); err != nil {
			app.Logger.Error("Watermill router stopped with an error", slog.String("error", err.Error()))
		}
	}()
}

func setupWatermill(
	leaderboardSvc leaderboardscoring.Service,
	subscriber message.Subscriber,
	config consumer.Config,
	logger *slog.Logger,
	redis *redis.Adapter,
) *message.Router {

	//watermillLogger := watermill.NewStdLogger(true, true)
	router, err := message.NewRouter(message.RouterConfig{}, nil) // Watermill's logger adapter can be added here.
	if err != nil {
		logger.Error("Failed to create Watermill router", slog.String("error", err.Error()))
		panic(err)
	}

	checker := consumer.NewIdempotencyChecker(redis.Client(), config, logger)
	handler := consumer.NewHandler(leaderboardSvc, checker, logger)

	router.AddNoPublisherHandler(
		"ContributionHandler",
		"CONTRIBUTION_REGISTERED",
		subscriber,
		handler.HandleContributionRegistered,
	)

	return router
}

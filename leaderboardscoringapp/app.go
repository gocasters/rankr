package leaderboardscoringapp

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer"
	leaderboardGRPC "github.com/gocasters/rankr/leaderboardscoringapp/delivery/grpc"
	leaderboardHTTP "github.com/gocasters/rankr/leaderboardscoringapp/delivery/http"
	postgrerepository "github.com/gocasters/rankr/leaderboardscoringapp/repository/database"
	"github.com/gocasters/rankr/leaderboardscoringapp/repository/queue"
	"github.com/gocasters/rankr/leaderboardscoringapp/repository/redisrepository"
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
	"time"
)

type Application struct {
	HTTPServer                leaderboardHTTP.Server
	LeaderboardGrpcServer     leaderboardGRPC.Server
	LeaderboardSvc            *leaderboardscoring.Service
	LeaderboardscoringHandler leaderboardHTTP.Handler
	WMRouter                  *message.Router
	WMLogger                  watermill.LoggerAdapter
	Config                    Config
	Subscriber                message.Subscriber
	RedisAdapter              *redis.Adapter
	DBConn                    *database.Database
	NatsAdapter               *nats.Adapter
	EventQueue                *queue.MemoryBatchQueue[leaderboardscoring.ProcessedScoreEvent]
	SnapshotQueue             *queue.MemoryBatchQueue[leaderboardscoring.UserTotalScore]
}

func Setup(ctx context.Context, config Config) *Application {
	logger := logger.L()

	if strings.TrimSpace(config.SubscriberTopic) == "" {
		logger.Error("SubscriberTopic is empty; set config.subscriber_topic")
		panic("invalid config: subscriber_topic")
	}

	// create database connection
	databaseConn, cnErr := database.Connect(config.PostgresDB)
	if cnErr != nil {
		logger.Error("fatal error occurred", "reason", "failed to connect to database", slog.Any("error", cnErr))
		return nil
	}

	// initial redis adapter
	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		databaseConn.Close()

		logger.Error("Failed to initialize Redis adapter", slog.String("error", err.Error()))
		panic(err)
	}

	// initial nats jetstream adapter
	wmLogger := watermill.NewStdLogger(true, true)
	natsAdapter, nErr := nats.New(ctx, config.Nats, wmLogger)
	if nErr != nil {
		databaseConn.Close()

		if cErr := redisAdapter.Close(); cErr != nil {
			logger.Error("Failed to Close redisAdapter.", slog.String("error", cErr.Error()))
		}

		logger.Error("Failed to create NATS adapter.", slog.String("error", nErr.Error()))
		panic(nErr)
	}

	// create subscriber
	subscriber := natsAdapter.Subscriber()

	// initial leaderboard-scoring repository, validator, service
	// init persistence, leaderboard, validator ...
	persistence := postgrerepository.NewPostgreSQLRepository(databaseConn, config.RetryConfig)
	leaderboard := redisrepository.NewRedisLeaderboardRepository(redisAdapter.Client())
	lbScoringValidator := leaderboardscoring.NewValidator()

	// build queues
	eventQueue := queue.NewMemoryBatchQueue[leaderboardscoring.ProcessedScoreEvent](
		100,
		5*time.Second,
		persistence.AddProcessedScoreEvents,
	)

	snapshotQueue := queue.NewMemoryBatchQueue[leaderboardscoring.UserTotalScore](
		100,
		30*time.Second,
		persistence.AddUserTotalScores,
	)

	lbScoringService := leaderboardscoring.NewService(
		persistence,
		leaderboard,
		eventQueue,
		snapshotQueue,
		lbScoringValidator,
	)

	// initial http server
	httpServer, hErr := httpserver.New(config.HTTPServer)
	if hErr != nil {
		logger.Error("Failed to initialize HTTP server", slog.String("error", hErr.Error()))
		panic(hErr)
	}
	lbScoringHandler := leaderboardHTTP.NewHandler()
	leaderboardHttpServer := leaderboardHTTP.New(httpServer, lbScoringHandler, lbScoringService)

	// initial rpc server
	rpcServer, gErr := grpc.NewServer(config.RPCServer)
	if gErr != nil {
		logger.Error("Failed to initialize gRPC server", slog.String("error", gErr.Error()))
		panic(gErr)
	}
	leaderboardGrpcHandler := leaderboardGRPC.NewHandler(lbScoringService)
	leaderboardGrpcServer := leaderboardGRPC.New(rpcServer, leaderboardGrpcHandler)

	// create one instance of Application
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
		DBConn:                    databaseConn,
		NatsAdapter:               natsAdapter,
		EventQueue:                eventQueue,
		SnapshotQueue:             snapshotQueue,
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

	router.AddConsumerHandler(
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

	var shutdownWg sync.WaitGroup
	shutdownWg.Add(4)

	go app.shutdownHTTPServer(ctx, &shutdownWg)
	go app.shutdownWatermillRouter(ctx, &shutdownWg)
	go app.shutdownGRPCServer(ctx, &shutdownWg)
	go app.shutdownResources(ctx, &shutdownWg)

	done := make(chan struct{})
	go func() {
		shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All servers have been shutdown successfully.")
		return true
	case <-ctx.Done():
		logger.Warn("Shutdown timed out, exiting application.")
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

func (app *Application) shutdownWatermillRouter(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("Closing Watermill router...")
	if err := app.WMRouter.Close(); err != nil {
		log.Error("Watermill router graceful shutdown failed", "error", err)
	} else {
		log.Info("Watermill router shutdown successfully.")
	}
}

func (app *Application) shutdownGRPCServer(parentCtx context.Context, wg *sync.WaitGroup) {
	logger := logger.L()
	defer wg.Done()
	logger.Info("starting gracefully shutdown leaderboard-scoring gRPC server")

	app.LeaderboardGrpcServer.Stop()

	logger.Info("leaderboard-scoring gRPC server shutdown successfully.")
}

func (app *Application) shutdownResources(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	log := logger.L()

	log.Info("Stop Event Queue...")
	app.EventQueue.Stop()

	log.Info("Stop Snapshot Queue...")
	app.SnapshotQueue.Stop()

	log.Info("Close NATS adapter...")
	if err := app.NatsAdapter.Close(); err != nil {
		log.Error("Failed to close NATS adapter.", slog.String("error", err.Error()))
	}

	log.Info("Close Database Connection...")
	app.DBConn.Close()

	log.Info("Close Redis adapter...")
	if err := app.RedisAdapter.Close(); err != nil {
		log.Error("Failed to close Redis adapter.", slog.String("error", err.Error()))
	}
}

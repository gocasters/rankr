package leaderboardscoringapp

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/natsadapter"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer/batchprocessor"
	"github.com/gocasters/rankr/leaderboardscoringapp/delivery/consumer/rawevent"
	leaderboardGRPC "github.com/gocasters/rankr/leaderboardscoringapp/delivery/grpc"
	leaderboardHTTP "github.com/gocasters/rankr/leaderboardscoringapp/delivery/http"
	postgrerepository "github.com/gocasters/rankr/leaderboardscoringapp/repository/database"
	"github.com/gocasters/rankr/leaderboardscoringapp/repository/redisrepository"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/topicsname"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	HTTPServer                leaderboardHTTP.Server
	LeaderboardGrpcServer     leaderboardGRPC.Server
	LeaderboardSvc            *leaderboardscoring.Service
	LeaderboardscoringHandler leaderboardHTTP.Handler
	WMRouter                  *message.Router
	WMLogger                  watermill.LoggerAdapter
	Config                    Config
	WMSubscriber              message.Subscriber
	RedisAdapter              *redis.Adapter
	DBConn                    *database.Database
	NatsWMAdapter             *nats.Adapter
	NatsAdapter               *natsadapter.Adapter
	BatchProcessor            *batchprocessor.Processor
}

func Setup(ctx context.Context, config Config) *Application {
	log := logger.L()

	config.StreamNameRawEvents = topicsname.StreamNameRawEvents

	// Initialize PostgreSQL connection
	databaseConn, err := database.Connect(config.PostgresDB)
	if err != nil {
		log.Error("failed to establish PostgreSQL connection",
			slog.String("error", err.Error()))
		panic(err)
	}
	log.Info("PostgreSQL connection established successfully")

	// Initialize Redis adapter
	redisAdapter, err := redis.New(ctx, config.Redis)
	if err != nil {
		databaseConn.Close()
		log.Error("failed to initialize Redis adapter",
			slog.String("error", err.Error()))
		panic(err)
	}
	log.Info("Redis adapter initialized successfully")

	// Initialize NATS Watermill adapter (for raw events consumption)
	wmLogger := watermill.NewStdLogger(true, true)
	natsWMAdapter, err := nats.New(ctx, config.WatermillNats, wmLogger)
	if err != nil {
		databaseConn.Close()
		_ = redisAdapter.Close()
		log.Error("failed to initialize NATS Watermill adapter",
			slog.String("error", err.Error()))
		panic(err)
	}
	log.Info("NATS Watermill adapter initialized successfully")

	// Initialize NATS native adapter (for processed events publishing/consumption)
	if config.NatsAdapter.StreamName == "" {
		config.NatsAdapter.StreamName = topicsname.StreamNameLeaderboardscoringProcessedEvents
	}
	if config.NatsAdapter.StreamSubjects == nil {
		config.NatsAdapter.StreamSubjects = []string{
			topicsname.TopicProcessedScoreEvents,
			topicsname.TopicProcessedScoreEventsDLQ,
		}
	}
	natsAdapter, err := natsadapter.New(config.NatsAdapter, log)
	if err != nil {
		databaseConn.Close()
		_ = redisAdapter.Close()
		_ = natsWMAdapter.Close()
		log.Error("failed to initialize NATS native adapter",
			slog.String("error", err.Error()))
		panic(err)
	}
	log.Info("NATS native adapter initialized successfully",
		slog.String("stream", config.NatsAdapter.StreamName))

	// Initialize repositories
	persistence := postgrerepository.NewPostgreSQLRepository(databaseConn, config.DatabaseRetry)
	leaderboard := redisrepository.NewRedisLeaderboardRepository(redisAdapter.Client())
	lbScoringValidator := leaderboardscoring.NewValidator()

	// Initialize leaderboard scoring service
	lbScoringService := leaderboardscoring.NewService(
		persistence,
		leaderboard,
		natsAdapter,
		topicsname.TopicProcessedScoreEvents,
		lbScoringValidator,
	)
	log.Info("leaderboard scoring service initialized")

	// Initialize HTTP server
	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		log.Error("failed to initialize HTTP server",
			slog.String("error", err.Error()))
		panic(err)
	}
	lbScoringHandler := leaderboardHTTP.NewHandler()
	leaderboardHttpServer := leaderboardHTTP.New(httpServer, lbScoringHandler, lbScoringService)

	// Initialize gRPC server
	rpcServer, err := grpc.NewServer(config.RPCServer)
	if err != nil {
		log.Error("failed to initialize gRPC server",
			slog.String("error", err.Error()))
		panic(err)
	}
	leaderboardGrpcHandler := leaderboardGRPC.NewHandler(lbScoringService)
	leaderboardGrpcServer := leaderboardGRPC.New(rpcServer, leaderboardGrpcHandler)

	// Create NATS pull consumer for batch processing
	pullConsumer, err := natsAdapter.CreatePullConsumer(config.PullConsumer)
	if err != nil {
		log.Error("failed to create NATS pull consumer",
			slog.String("error", err.Error()),
			slog.String("consumer", config.PullConsumer.DurableName))
		panic(err)
	}
	log.Info("NATS pull consumer created successfully",
		slog.String("consumer", config.PullConsumer.DurableName),
		slog.Int("batch_size", config.PullConsumer.BatchSize))

	// Initialize batch processor
	processor := batchprocessor.NewProcessor(
		pullConsumer,
		natsAdapter,
		persistence,
		config.BatchProcessor,
	)
	log.Info("batch processor initialized",
		slog.Duration("tick_interval", config.BatchProcessor.TickInterval),
		slog.Duration("metrics_interval", config.BatchProcessor.MetricsInterval))

	return &Application{
		HTTPServer:                leaderboardHttpServer,
		LeaderboardGrpcServer:     leaderboardGrpcServer,
		LeaderboardSvc:            lbScoringService,
		LeaderboardscoringHandler: lbScoringHandler,
		WMRouter:                  nil,
		WMLogger:                  wmLogger,
		Config:                    config,
		WMSubscriber:              natsWMAdapter.Subscriber(),
		RedisAdapter:              redisAdapter,
		DBConn:                    databaseConn,
		NatsWMAdapter:             natsWMAdapter,
		NatsAdapter:               natsAdapter,
		BatchProcessor:            processor,
	}
}

func (app *Application) Start() {
	log := logger.L()
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app.startHTTPServer(&wg)
	app.startWatermill(ctx, &wg)
	app.startGRPCServer(&wg)
	app.startBatchProcessor(ctx, &wg)

	log.Info("leaderboard scoring application is ready and running")

	<-ctx.Done()
	log.Info("shutdown signal received, initiating graceful shutdown")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.Shutdown(shutdownTimeoutCtx) {
		log.Info("application shutdown completed successfully")
	} else {
		log.Warn("shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}

	wg.Wait()
	log.Info("leaderboard scoring application stopped")
}

func (app *Application) startHTTPServer(wg *sync.WaitGroup) {
	log := logger.L()
	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Info("starting HTTP server",
			slog.String("host", app.Config.HTTPServer.Host),
			slog.Int("port", app.Config.HTTPServer.Port))

		if err := app.HTTPServer.Serve(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("HTTP server failed",
					slog.Int("port", app.Config.HTTPServer.Port),
					slog.String("error", err.Error()))
				panic(err)
			}
		}

		log.Info("HTTP server stopped")
	}()
}

func (app *Application) startWatermill(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	app.setupWatermill()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Info("starting Watermill event router")

		if err := app.WMRouter.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("Watermill router failed",
				slog.String("error", err.Error()))
			panic(err)
		}

		log.Info("Watermill event router stopped")
	}()
}

func (app *Application) setupWatermill() {
	log := logger.L()

	router, err := message.NewRouter(message.RouterConfig{}, app.WMLogger)
	if err != nil {
		log.Error("failed to create Watermill router",
			slog.String("error", err.Error()))
		panic(err)
	}

	checker := rawevent.NewIdempotencyChecker(app.RedisAdapter.Client(), app.Config.RawEventConsumer)
	rawEventHandler := rawevent.NewHandler(app.LeaderboardSvc, checker)

	router.AddConsumerHandler(
		"RawEventHandler",
		app.Config.StreamNameRawEvents,
		app.WMSubscriber,
		rawEventHandler.HandleEvent,
	)

	log.Info("Watermill router configured",
		slog.String("topic", app.Config.StreamNameRawEvents))

	app.WMRouter = router
}

func (app *Application) startGRPCServer(wg *sync.WaitGroup) {
	log := logger.L()
	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Info("starting gRPC server",
			slog.Int("port", app.Config.RPCServer.Port))

		if err := app.LeaderboardGrpcServer.Serve(); err != nil {
			log.Error("gRPC server failed",
				slog.String("error", err.Error()))
		}

		log.Info("gRPC server stopped")
	}()
}

func (app *Application) startBatchProcessor(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Info("starting batch processor worker")

		if err := app.BatchProcessor.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error("batch processor failed",
					slog.String("error", err.Error()))
			}
		}

		log.Info("batch processor worker stopped")
	}()
}

func (app *Application) Shutdown(ctx context.Context) bool {
	log := logger.L()
	log.Info("initiating graceful shutdown sequence")

	var shutdownWg sync.WaitGroup
	shutdownWg.Add(5)

	go app.shutdownHTTPServer(ctx, &shutdownWg)
	go app.shutdownWatermillRouter(ctx, &shutdownWg)
	go app.shutdownGRPCServer(ctx, &shutdownWg)
	go app.shutdownBatchProcessor(&shutdownWg)
	go app.shutdownResources(ctx, &shutdownWg)

	done := make(chan struct{})
	go func() {
		shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		log.Warn("shutdown timeout reached")
		return false
	}
}

func (app *Application) shutdownHTTPServer(parentCtx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("shutting down HTTP server")

	httpCtx, cancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer cancel()

	if err := app.HTTPServer.HTTPServer.Stop(httpCtx); err != nil {
		log.Error("HTTP server shutdown failed",
			slog.String("error", err.Error()))
	} else {
		log.Info("HTTP server shutdown complete")
	}
}

func (app *Application) shutdownWatermillRouter(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("shutting down Watermill router")

	if err := app.WMRouter.Close(); err != nil {
		log.Error("Watermill router shutdown failed",
			slog.String("error", err.Error()))
	} else {
		log.Info("Watermill router shutdown complete")
	}
}

func (app *Application) shutdownGRPCServer(parentCtx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("shutting down gRPC server")
	app.LeaderboardGrpcServer.Stop()
	log.Info("gRPC server shutdown complete")
}

func (app *Application) shutdownBatchProcessor(wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("shutting down batch processor")

	if err := app.BatchProcessor.Close(); err != nil {
		log.Error("batch processor shutdown failed",
			slog.String("error", err.Error()))
	} else {
		log.Info("batch processor shutdown complete")
	}
}

func (app *Application) shutdownResources(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("closing NATS Watermill adapter")
	if err := app.NatsWMAdapter.Close(); err != nil {
		log.Error("failed to close NATS Watermill adapter",
			slog.String("error", err.Error()))
	}

	log.Info("closing NATS native adapter")
	if err := app.NatsAdapter.Close(); err != nil {
		log.Error("failed to close NATS native adapter",
			slog.String("error", err.Error()))
	}

	log.Info("closing PostgreSQL connection")
	app.DBConn.Close()

	log.Info("closing Redis adapter")
	if err := app.RedisAdapter.Close(); err != nil {
		log.Error("failed to close Redis adapter",
			slog.String("error", err.Error()))
	}

	log.Info("all resources released")
}

package taskapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/topicsname"
	"github.com/gocasters/rankr/taskapp/delivery/consumer"
	"github.com/gocasters/rankr/taskapp/delivery/http"
	"github.com/gocasters/rankr/taskapp/repository"
	"github.com/gocasters/rankr/taskapp/service/task"
)

type Application struct {
	TaskRepo     task.Repository
	TaskSrv      *task.Service
	TaskHandler  http.Handler
	HTTPServer   http.Server
	WMRouter     *message.Router
	WMLogger     watermill.LoggerAdapter
	WMSubscriber message.Subscriber
	NatsAdapter  *nats.Adapter
	Config       Config
	Logger       *slog.Logger
}

func Setup(
	ctx context.Context,
	config Config,
	postgresConn *database.Database,
	logger *slog.Logger,
) (Application, error) {

	if config.StreamNameRawEvents == "" {
		config.StreamNameRawEvents = topicsname.StreamNameRawEvents
	}

	taskRepo := repository.NewTaskRepo(config.Repository, postgresConn, logger)
	taskValidator := task.NewValidator(taskRepo)
	taskSvc := task.NewService(taskRepo, taskValidator, logger)

	taskHandler := http.NewHandler(taskSvc, logger)

	server, err := httpserver.New(config.HTTPServer)
	if err != nil {
		logger.Error("failed to initialize HTTP server", "err", err)
		return Application{}, err
	}

	httpServer := http.New(
		*server,
		taskHandler,
		logger,
	)

	wmLogger := watermill.NewStdLogger(true, true)
	natsAdapter, err := nats.New(ctx, config.WatermillNats, wmLogger)
	if err != nil {
		logger.Error("failed to initialize NATS Watermill adapter", slog.String("error", err.Error()))
		return Application{}, err
	}
	logger.Info("NATS Watermill adapter initialized successfully")

	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		logger.Error("failed to initialize Watermill router", slog.String("error", err.Error()))
		return Application{}, err
	}

	eventHandler := consumer.NewHandler(&taskSvc)

	subscriber := natsAdapter.Subscriber()

	router.AddConsumerHandler(
		"task_event_consumer",
		config.StreamNameRawEvents,
		subscriber,
		eventHandler.HandleEvent,
	)

	logger.Info("Task event consumer registered",
		slog.String("stream", config.StreamNameRawEvents),
	)

	return Application{
		TaskRepo:     taskRepo,
		TaskSrv:      &taskSvc,
		TaskHandler:  taskHandler,
		HTTPServer:   httpServer,
		WMRouter:     router,
		WMLogger:     wmLogger,
		WMSubscriber: subscriber,
		NatsAdapter:  natsAdapter,
		Config:       config,
		Logger:       logger,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	<-ctx.Done()
	app.Logger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		app.Logger.Info("Servers shut down gracefully")
	} else {
		app.Logger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	app.Logger.Info("task_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), slog.Any("error", err))
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info("Event consumer started")
		if err := app.WMRouter.Run(context.Background()); err != nil {
			app.Logger.Error("Event consumer stopped with error", slog.Any("error", err))
		} else {
			app.Logger.Info("Event consumer stopped gracefully")
		}
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	parentCtx := ctx
	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(2)
		go app.shutdownHTTPServer(parentCtx, &shutdownWg)
		go app.shutdownEventConsumer(&shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
		app.Logger.Info("All servers have been shut down successfully.")

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
	httpShutdownCtx, httpCancel := context.WithTimeout(parentCtx, app.Config.HTTPServer.ShutdownTimeout)
	defer httpCancel()
	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		app.Logger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}

	app.Logger.Info("HTTP server shut down successfully.")
}

func (app Application) shutdownEventConsumer(wg *sync.WaitGroup) {
	app.Logger.Info("Starting graceful shutdown for event consumer")

	defer wg.Done()
	if err := app.WMRouter.Close(); err != nil {
		app.Logger.Error("Event consumer shutdown failed", slog.Any("error", err))
	} else {
		app.Logger.Info("Event consumer shut down successfully")
	}

	if err := app.NatsAdapter.Close(); err != nil {
		app.Logger.Error("NATS adapter close failed", slog.Any("error", err))
	}
}

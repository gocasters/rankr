package realtimeapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	natsadapter "github.com/gocasters/rankr/adapter/nats"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/realtimeapp/constant"
	"github.com/gocasters/rankr/realtimeapp/delivery/http"
	"github.com/gocasters/rankr/realtimeapp/repository"
	"github.com/gocasters/rankr/realtimeapp/service/realtime"
)

type Application struct {
	ConnectionStore *repository.ConnectionStore
	RealtimeSrv     realtime.Service
	RealtimeHandler http.Handler
	HTTPServer      http.Server
	NATSSubscriber  *realtime.Subscriber
	NATSAdapter     *natsadapter.Adapter
	Config          Config
	Logger          *slog.Logger
	Redis           *redis.Adapter
}

func Setup(
	ctx context.Context,
	config Config,
	logger *slog.Logger,
) (Application, error) {

	var redisAdapter *redis.Adapter
	var err error

	if config.Redis.Host != "" {
		redisAdapter, err = redis.New(ctx, config.Redis)
		if err != nil {
			logger.Warn("failed to initialize Redis, continuing without it", "err", err)
		}
	} else {
		logger.Info("Redis not configured, skipping")
	}

	connectionStore := repository.NewConnectionStore(logger)

	realtimeService := realtime.NewService(connectionStore, logger)

	realtimeHandler := http.NewHandler(realtimeService, logger)

	httpServer, err := httpserver.New(config.HTTPServer)
	if err != nil {
		logger.Error("failed to initialize HTTP server", "err", err)
		return Application{}, err
	}

	var natsAdapter *natsadapter.Adapter
	var natsSubscriber *realtime.Subscriber
	if config.NATS.URL != "" {
		watermillLogger := watermill.NewSlogLogger(logger)
		natsAdapter, err = natsadapter.New(ctx, config.NATS, watermillLogger)
		if err != nil {
			logger.Warn("failed to initialize NATS adapter, continuing without event streaming", "err", err)
		} else {
			topics := parseTopics(config.SubscribeTopics)
			natsSubscriber = realtime.NewSubscriber(natsAdapter.Subscriber(), realtimeService, topics, logger)
		}
	} else {
		logger.Info("NATS not configured, event streaming disabled")
	}

	return Application{
		ConnectionStore: connectionStore,
		RealtimeSrv:     realtimeService,
		RealtimeHandler: realtimeHandler,
		HTTPServer: http.New(
			*httpServer,
			realtimeHandler,
			logger,
		),
		NATSSubscriber: natsSubscriber,
		NATSAdapter:    natsAdapter,
		Config:         config,
		Logger:         logger,
		Redis:          redisAdapter,
	}, nil
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if app.NATSSubscriber != nil {
		if err := app.NATSSubscriber.Start(ctx); err != nil {
			app.Logger.Error("failed to start NATS subscriber", "error", err)
		}
	}

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
	app.Logger.Info("realtime_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("✅ HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), "error", err)
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	app.Logger.Info("Starting server shutdown process...")
	shutdownDone := make(chan struct{})

	parentCtx := context.Background()
	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(parentCtx, &shutdownWg)

		if app.NATSSubscriber != nil {
			shutdownWg.Add(1)
			go func() {
				defer shutdownWg.Done()
				if err := app.NATSSubscriber.Stop(); err != nil {
					app.Logger.Error("failed to stop NATS subscriber", "error", err)
				}
			}()
		}

		if app.NATSAdapter != nil {
			shutdownWg.Add(1)
			go func() {
				defer shutdownWg.Done()
				if err := app.NATSAdapter.Close(); err != nil {
					app.Logger.Error("failed to close NATS adapter", "error", err)
				}
			}()
		}

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

func parseTopics(topicsStr string) []string {
	if topicsStr == "" {

		return []string{
			constant.TopicContributorCreated,
			constant.TopicContributorUpdated,
			constant.TopicTaskCreated,
			constant.TopicTaskUpdated,
			constant.TopicTaskCompleted,
			constant.TopicLeaderboardScored,
			constant.TopicLeaderboardUpdated,
			constant.TopicProjectCreated,
			constant.TopicProjectUpdated,
		}
	}

	topics := strings.Split(topicsStr, ",")
	for i := range topics {
		topics[i] = strings.TrimSpace(topics[i])
	}
	return topics
}

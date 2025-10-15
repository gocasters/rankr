package batchprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/nats-io/nats.go"
	"log/slog"
	"time"
)

type PullConsumer interface {
	Fetch() ([]*nats.Msg, error)
	GetConsumerInfo() (*nats.ConsumerInfo, error)
	Close() error
}

type Config struct {
	TickInterval    time.Duration `koanf:"tick_interval"`    // TickInterval How often to check for messages (e.g., 1s)
	MetricsInterval time.Duration `koanf:"metrics_interval"` // MetricsInterval How often to log metrics (e.g., 30s)
}
type Processor struct {
	consumer    PullConsumer
	persistence leaderboardscoring.EventPersistence
	config      Config
}

func NewProcessor(
	consumer PullConsumer,
	persistence leaderboardscoring.EventPersistence,
	config Config,
) *Processor {
	return &Processor{
		consumer:    consumer,
		persistence: persistence,
		config:      config,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	var defaultTickInterval = 20 * time.Second
	var defaultMetricsInterval = 30 * time.Second

	log := logger.L()
	log.Info("Starting batch processor")

	if p.config.TickInterval <= 0 {
		log.Warn(fmt.Sprintf("invalid tick interval: must be > 0, got %s", p.config.TickInterval),
			slog.String("default set", "20s"))
		p.config.TickInterval = defaultTickInterval
	}

	if p.config.MetricsInterval <= 0 {
		log.Warn(fmt.Sprintf("invalid metrics interval: must be > 0, got %s", p.config.MetricsInterval),
			slog.String("default set", "30s"))
		p.config.MetricsInterval = defaultMetricsInterval
	}

	ticker := time.NewTicker(p.config.TickInterval)
	defer ticker.Stop()

	metricsTicker := time.NewTicker(p.config.MetricsInterval)
	defer metricsTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Batch processor stopped by context")
			return ctx.Err()

		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				log.Error("Failed to process batch", slog.String("error", err.Error()))
			}

		case <-metricsTicker.C:
			p.logMetrics()
		}
	}
}

// Close gracefully shuts down the processor
func (p *Processor) Close() error {
	return p.consumer.Close()
}

// processBatch fetches and processes a single batch
// 1. Fetch messages
// 2. Parse messages
// 3. Terminate invalid messages (don't retry)
// 4. Persist valid events
// 5. ACK all valid messages on success
func (p *Processor) processBatch(ctx context.Context) error {
	log := logger.L()

	// Fetch messages
	msgs, err := p.consumer.Fetch()
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}

	if len(msgs) == 0 {
		return nil
	}

	log.Debug("Fetched messages", slog.Int("count", len(msgs)))

	// Parse messages
	events := make([]leaderboardscoring.ProcessedScoreEvent, 0, len(msgs))
	validMsgs := make([]*nats.Msg, 0, len(msgs))
	invalidMsgs := make([]*nats.Msg, 0)

	for _, msg := range msgs {
		var event leaderboardscoring.ProcessedScoreEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Warn(
				"Invalid message format",
				slog.String("error", err.Error()),
				slog.String("subject", msg.Subject),
			)
			invalidMsgs = append(invalidMsgs, msg)
			continue
		}

		validMsgs = append(validMsgs, msg)
		events = append(events, event)
	}

	// Terminate invalid messages (don't retry)
	/*for _, invalidMsg := range invalidMsgs {
		if err := invalidMsg.Term(); err != nil {
			log.Error("Failed to terminate invalid message",
				slog.String("error", err.Error()))
		}
	}*/

	if len(events) == 0 {
		log.Debug("No valid messages in batch")
		return nil
	}

	// Persist valid events
	log.Info("Persisting batch", slog.Int("count", len(events)))

	if err := p.persistence.AddProcessedScoreEvents(ctx, events); err != nil {
		log.Error(
			"Failed to persist batch",
			slog.Int("count", len(events)),
			slog.String("error", err.Error()),
		)

		// NAK all valid messages for retry
		for _, validMsg := range validMsgs {
			if nakErr := validMsg.Nak(); nakErr != nil {
				log.Error(
					"Failed to NAK message",
					slog.String("error", nakErr.Error()),
				)
			}
		}

		return fmt.Errorf("persist events: %w", err)
	}

	// ACK all valid messages on success
	for _, validMsg := range validMsgs {
		if err := validMsg.Ack(); err != nil {
			log.Error(
				"Failed to ACK message",
				slog.String("error", err.Error()),
			)
		}
	}

	log.Info(
		"Batch processed successfully",
		slog.Int("persisted", len(events)),
		slog.Int("invalid", len(invalidMsgs)),
	)

	return nil
}

// logMetrics logs consumer statistics
func (p *Processor) logMetrics() {
	info, err := p.consumer.GetConsumerInfo()
	if err != nil {
		logger.L().Error(
			"Failed to get consumer info",
			slog.String("error", err.Error()),
		)
		return
	}

	logger.L().Info(
		"Consumer processed events metrics.",
		slog.Uint64("pending", info.NumPending),
		slog.Int("ack_pending", info.NumAckPending),
		slog.Int("redelivered", info.NumRedelivered),
		slog.Int("waiting", info.NumWaiting),
	)
}

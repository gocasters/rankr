package recovery

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service/delivery"
)

const (
	contextTimeoutFactor = 0.8
)

type GithubClient interface {
	GetDeliveries(webhookConfig WebhookConfig, page int, perPage int) ([]WebhookDelivery, error)
	ReattemptDelivery(webhookConfig WebhookConfig, deliveryID string) error
}

type WebhookConfig struct {
	Repo   string `koanf:"repo"`
	Owner  string `koanf:"owner"`
	HookID string `koanf:"hook_id"`
	Token  string `koanf:"token"`
}

type Config struct {
	RecoveryLostDeliveriesIntervalInSeconds int             `koanf:"recovery_lost_deliveries_interval_in_seconds"`
	BatchSize                               int             `koanf:"batch_size"`
	DeliveryPerPage                         int             `koanf:"delivery_per_page"`
	Webhooks                                []WebhookConfig `koanf:"webhooks"`
}

type LostDeliveriesScheduler struct {
	config       Config
	webhookSvc   delivery.Service
	githubClient GithubClient
	scheduler    *gocron.Scheduler
	mu           sync.RWMutex
	running      bool
}

type recoveryStats struct {
	totalRecovered int
	totalFailed    int
	webhookStats   map[string]*webhookStats
}

type webhookStats struct {
	recovered int
	failed    int
	errors    []error
}

func NewSchedulerService(config Config, webhookSvc delivery.Service, githubClient GithubClient) *LostDeliveriesScheduler {
	return &LostDeliveriesScheduler{
		config:       config,
		webhookSvc:   webhookSvc,
		githubClient: githubClient,
		scheduler:    gocron.NewScheduler(time.UTC),
	}
}

func (s *LostDeliveriesScheduler) Start(done <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.L().Warn("scheduler already running, skipping start")
		return
	}
	s.running = true
	s.mu.Unlock()

	_, err := s.scheduler.Every(s.config.RecoveryLostDeliveriesIntervalInSeconds).Seconds().Do(s.RunDeliveryCheck)
	if err != nil {
		logger.L().Error("failed to schedule delivery check", "error", err)
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return
	}

	s.scheduler.StartAsync()
	logger.L().Info("scheduler started", "interval_seconds", s.config.RecoveryLostDeliveriesIntervalInSeconds)

	<-done
	logger.L().Info("stopping scheduler")
	s.scheduler.Stop()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

func (s *LostDeliveriesScheduler) RunDeliveryCheck() {
	logger.L().Info("starting scheduled delivery check")
	startTime := time.Now()

	contextTimeout := s.calculateContextTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	stats := s.processAllWebhooks(ctx)

	duration := time.Since(startTime)
	s.logCompletionStats(stats, duration)
}

func (s *LostDeliveriesScheduler) calculateContextTimeout() time.Duration {
	timeoutSeconds := int(math.Ceil(contextTimeoutFactor * float64(s.config.RecoveryLostDeliveriesIntervalInSeconds)))
	return time.Duration(timeoutSeconds) * time.Second
}

func (s *LostDeliveriesScheduler) processAllWebhooks(ctx context.Context) *recoveryStats {
	stats := &recoveryStats{
		webhookStats: make(map[string]*webhookStats),
	}

	for _, webhook := range s.config.Webhooks {
		if ctx.Err() != nil {
			logger.L().Warn("context cancelled, stopping webhook processing")
			break
		}

		webhookKey := fmt.Sprintf("%s/%s", webhook.Owner, webhook.Repo)
		webhookStat := s.checkWebhookDeliveries(ctx, webhook)

		stats.webhookStats[webhookKey] = webhookStat
		stats.totalRecovered += webhookStat.recovered
		stats.totalFailed += webhookStat.failed
	}

	return stats
}

func (s *LostDeliveriesScheduler) checkWebhookDeliveries(ctx context.Context, webhook WebhookConfig) *webhookStats {
	stats := &webhookStats{
		errors: make([]error, 0),
	}

	webhookKey := fmt.Sprintf("%s/%s", webhook.Owner, webhook.Repo)
	logger.L().Info("checking webhook deliveries", "webhook", webhookKey)

	page := 1
	for {
		if ctx.Err() != nil {
			stats.errors = append(stats.errors, fmt.Errorf("context cancelled at page %d: %w", page, ctx.Err()))
			break
		}

		githubDeliveries, err := s.githubClient.GetDeliveries(webhook, page, s.config.DeliveryPerPage)
		if err != nil {
			stats.errors = append(stats.errors, fmt.Errorf("failed to get deliveries page %d: %w", page, err))
			break
		}

		// No more deliveries to process
		if len(githubDeliveries) == 0 {
			break
		}

		s.processDeliveryBatch(ctx, webhook, githubDeliveries, stats)

		// If we got fewer results than requested, we've reached the end
		if len(githubDeliveries) < s.config.DeliveryPerPage {
			break
		}

		page++
	}

	s.logWebhookStats(webhookKey, stats)
	return stats
}

func (s *LostDeliveriesScheduler) processDeliveryBatch(
	ctx context.Context,
	webhook WebhookConfig,
	githubDeliveries []WebhookDelivery,
	stats *webhookStats,
) {
	deliveryIDs := make([]string, 0, len(githubDeliveries))
	for _, deliveryItem := range githubDeliveries {
		deliveryIDs = append(deliveryIDs, deliveryItem.GUID)
	}

	lostDeliveries, err := s.webhookSvc.GetLostDeliveries(ctx, eventpb.EventProvider_EVENT_PROVIDER_GITHUB, deliveryIDs)
	if err != nil {
		stats.errors = append(stats.errors, fmt.Errorf("failed to get lost deliveries: %w", err))
		return
	}

	if len(lostDeliveries) == 0 {
		return
	}

	logger.L().Info("found lost deliveries", "count", len(lostDeliveries))

	for _, lostDeliveryID := range lostDeliveries {
		if ctx.Err() != nil {
			stats.errors = append(stats.errors, fmt.Errorf("context cancelled during reattempt: %w", ctx.Err()))
			return
		}

		err := s.githubClient.ReattemptDelivery(webhook, lostDeliveryID)
		if err != nil {
			stats.failed++
			stats.errors = append(stats.errors, fmt.Errorf("failed to reattempt delivery %s: %w", lostDeliveryID, err))
			// Continue processing other deliveries instead of breaking
			continue
		}
		stats.recovered++
	}
}

func (s *LostDeliveriesScheduler) logWebhookStats(webhookKey string, stats *webhookStats) {
	if stats.recovered == 0 && stats.failed == 0 {
		logger.L().Info("no lost deliveries found", "webhook", webhookKey)
		return
	}

	logFields := []interface{}{
		"webhook", webhookKey,
		"recovered", stats.recovered,
	}

	if stats.failed > 0 {
		logFields = append(logFields, "failed", stats.failed)
	}

	if len(stats.errors) > 0 {
		// Log first few errors for visibility
		errorCount := len(stats.errors)
		maxErrorsToLog := 3
		if errorCount > maxErrorsToLog {
			logFields = append(logFields, "error_count", errorCount)
			logFields = append(logFields, "sample_errors", stats.errors[:maxErrorsToLog])
		} else {
			logFields = append(logFields, "errors", stats.errors)
		}
	}

	logger.L().Info("webhook delivery check completed", logFields...)
}

func (s *LostDeliveriesScheduler) logCompletionStats(stats *recoveryStats, duration time.Duration) {
	logFields := []interface{}{
		"total_recovered", stats.totalRecovered,
		"total_failed", stats.totalFailed,
		"webhooks_processed", len(stats.webhookStats),
		"duration_seconds", duration.Seconds(),
	}

	if stats.totalRecovered == 0 && stats.totalFailed == 0 {
		logger.L().Info("scheduled delivery check completed - no recoveries needed", logFields...)
	} else {
		logger.L().Info("scheduled delivery check completed", logFields...)
	}
}

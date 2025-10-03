// scheduler TODO: This scheduler is currently implemented using a simple time.Ticker for the MVP.
// For more advanced scheduling capabilities (e.g., cron expressions), this should be
// refactored in the future to use a more robust library like gocron or robfig/cron.

package scheduler

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"log/slog"
	"sync"
	"time"
)

type Config struct {
	TickerInterval time.Duration `koanf:"ticker_interval"`
}

// JobScheduler triggers processing jobs periodically or based on queue size
type JobScheduler struct {
	service      *leaderboardscoring.Service
	queue        leaderboardscoring.EventQueue
	cfg          Config
	maxQueueSize int
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

func NewJobScheduler(
	service *leaderboardscoring.Service,
	queue leaderboardscoring.EventQueue,
	cfg Config,
	maxQueueSize int,
) *JobScheduler {
	return &JobScheduler{
		service:      service,
		queue:        queue,
		cfg:          cfg,
		maxQueueSize: maxQueueSize,
		stopCh:       make(chan struct{}),
	}
}

func (j *JobScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(j.cfg.TickerInterval)
	sizeTicker := time.NewTicker(1 * time.Second)
	log := logger.L()

	j.wg.Add(2)

	go func() {
		defer j.wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := j.service.ProcessEventQueue(ctx); err != nil {
					log.Error("Ticker-based queue processing failed", slog.String("error", err.Error()))
				}

			case <-j.stopCh:
				return
			}
		}
	}()

	// monitor queue size asynchronously
	go func() {
		defer j.wg.Done()
		defer sizeTicker.Stop()

		for {
			select {
			case <-sizeTicker.C:
				if j.queue.Size() >= j.maxQueueSize {
					if err := j.service.ProcessEventQueue(ctx); err != nil {
						log.Error("Size-based queue processing failed", slog.String("error", err.Error()))
					}
				}

			case <-j.stopCh:
				return
			}
		}
	}()

	log.Info("JobScheduler is started...")
}

// Stop stops the scheduler
func (j *JobScheduler) Stop() {
	logger.L().Info("Stop Scheduler...")

	close(j.stopCh)
	j.wg.Wait()
}

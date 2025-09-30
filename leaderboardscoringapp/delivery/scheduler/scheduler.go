// scheduler TODO: This scheduler is currently implemented using a simple time.Ticker for the MVP.
// For more advanced scheduling capabilities (e.g., cron expressions), this should be
// refactored in the future to use a more robust library like gocron or robfig/cron.

package scheduler

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
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

	go func() {
		for {
			select {
			case <-ticker.C:
				_ = j.service.ProcessEventQueue(ctx)
			case <-j.stopCh:
				ticker.Stop()
				return
			}
		}
	}()

	// monitor queue size asynchronously
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				if j.queue.Size() >= j.maxQueueSize {
					_ = j.service.ProcessEventQueue(ctx)
				}
			case <-j.stopCh:
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (j *JobScheduler) Stop() {
	close(j.stopCh)
}

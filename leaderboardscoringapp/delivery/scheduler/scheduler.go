// scheduler TODO: This scheduler is currently implemented using a simple time.Ticker for the MVP.
// For more advanced scheduling capabilities (e.g., cron expressions), this should be
// refactored in the future to use a more robust library like gocron or robfig/cron.

package scheduler

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"log/slog"
	"time"
)

type Config struct {
	Interval time.Duration `koanf:"interval"`
}
type Scheduler struct {
	leaderboardScoringSvc leaderboardscoring.Service
	cfg                   Config
	logger                *slog.Logger
	stopChan              chan struct{}
}

func New(lbScoringSvc leaderboardscoring.Service, logger *slog.Logger, cfg Config) *Scheduler {
	return &Scheduler{
		leaderboardScoringSvc: lbScoringSvc,
		cfg:                   cfg,
		logger:                logger,
		stopChan:              make(chan struct{}),
	}
}

// Start begins the periodic execution of the persistence job.
func (s *Scheduler) Start(ctx context.Context) {
	if s.cfg.Interval <= 0 {
		s.logger.Error("scheduler not started: non-positive interval", slog.Duration("interval", s.cfg.Interval))
		return
	}
	ticker := time.NewTicker(s.cfg.Interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.logger.Debug("Scheduler ticked. Processing persistence queue...")
				if err := s.leaderboardScoringSvc.ProcessPersistenceQueue(ctx); err != nil {
					s.logger.Error("Error processing persistence queue", slog.String("error", err.Error()))
				}
			case <-s.stopChan:
				ticker.Stop()
				s.logger.Info("Persistence scheduler stopped.")
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

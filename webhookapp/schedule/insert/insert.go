package insert

import (
	"context"
	"github.com/go-co-op/gocron"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/webhookapp/service/delivery"
	"log/slog"
	"sync"
	"time"
)

type Config struct {
	BulkInsertIntervalInSeconds int `koanf:"bulk_insert_interval_in_seconds"`
}

type BulkInsertScheduler struct {
	config     Config
	webhookSvc delivery.Service
	scheduler  *gocron.Scheduler
}

func NewSchedulerService(config Config, webhookSvc delivery.Service) *BulkInsertScheduler {
	return &BulkInsertScheduler{
		config:     config,
		webhookSvc: webhookSvc,
		scheduler:  gocron.NewScheduler(time.UTC),
	}
}

func (s *BulkInsertScheduler) Start(done <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	s.scheduler.Every(s.config.BulkInsertIntervalInSeconds).Seconds().Do(s.BulkInsertEvents)
	s.scheduler.StartAsync()

	<-done
	logger.L().Info("stop delete expired posts scheduler..")
	s.scheduler.Stop()
}

func (s *BulkInsertScheduler) BulkInsertEvents() {
	defer func() {
		if r := recover(); r != nil {
			logger.L().Error("Panic recovered in bulk insert job",
				slog.Any("recovery", r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.webhookSvc.ProcessBatch(ctx)
	if err != nil {
		logger.L().Error("Bulk insert job failed",
			slog.Any("error", err),
			slog.String("note", "job will retry on next schedule"))
		return
	}

	logger.L().Info("✅ Bulk insert job completed successfully")
}

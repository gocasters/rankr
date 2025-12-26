package scheduler

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/logger"
	"log/slog"
	"sync"
	"time"
)

type Config struct {
	DailyScoreCalculationCron string        `koanf:"daily_score_calculation_cron"`
	PublicLeaderboardCron     string        `koanf:"public_leaderboard_cron"`
	JobContextTimeout         time.Duration `koanf:"job_context_timeout"`
}

type Scheduler struct {
	sch                gocron.Scheduler
	leaderboardStatSvc *leaderboardstat.Service
	cfg                Config
}

func New(leaderboardStatSvc *leaderboardstat.Service, schedulerCfg Config) Scheduler {
	sch, err := gocron.NewScheduler(gocron.WithLocation(time.Local))
	if err != nil {
		logger.L().Error("failed to create Scheduler", slog.String("error", err.Error()))
		panic(err)
	}

	if schedulerCfg.PublicLeaderboardCron == "" {
		schedulerCfg.PublicLeaderboardCron = "*/3 * * * *"
	}
	if schedulerCfg.JobContextTimeout <= 0 {
		schedulerCfg.JobContextTimeout = 3 * time.Minute
	}
	if schedulerCfg.DailyScoreCalculationCron == "" {
		schedulerCfg.DailyScoreCalculationCron = "0 2 * * *"
	}

	return Scheduler{
		sch:                sch,
		leaderboardStatSvc: leaderboardStatSvc,
		cfg:                schedulerCfg,
	}
}

func (s *Scheduler) Start(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("leaderboardstat scheduler started")

	if err := s.dailyScoreCalculationJob(ctx); err != nil {
		log.Error("failed to create daily score calculation job", slog.String("error", err.Error()))
	}

	if err := s.publicLeaderboardJob(ctx); err != nil {
		log.Error("failed to create public leaderboard job", slog.String("error", err.Error()))
	}

	s.sch.Start()

	<-ctx.Done()

	log.Info("leaderboardstat scheduler exiting...")

	if sErr := s.sch.Shutdown(); sErr != nil {
		log.Warn("cannot successfully shutdown scheduler", slog.String("error", sErr.Error()))
	}
}

func (s *Scheduler) dailyScoreCalculationJob(parentCtx context.Context) error {
	log := logger.L()

	if s.cfg.DailyScoreCalculationCron == "" {
		// Default to run daily at 2 AM
		s.cfg.DailyScoreCalculationCron = "0 2 * * *"
	}

	dailyJob, err := s.sch.NewJob(
		gocron.CronJob(s.cfg.DailyScoreCalculationCron, false),
		gocron.NewTask(func() { s.dailyScoreCalculationTask(parentCtx) }),
		gocron.WithSingletonMode(gocron.LimitModeWait),
		gocron.WithName("daily-score-calculation"),
		gocron.WithTags("leaderboardstat-service"),
	)
	if err != nil {
		return fmt.Errorf("failed to create daily score calculation job: %w", err)
	}

	log.Info("dailyScoreCalculation job created",
		slog.String("name", dailyJob.Name()),
		slog.String("uuid", dailyJob.ID().String()),
		slog.Any("tags", dailyJob.Tags()),
		slog.String("crontab", s.cfg.DailyScoreCalculationCron),
	)

	return nil
}

func (s *Scheduler) dailyScoreCalculationTask(parentCtx context.Context) {
	log := logger.L()

	log.Info("dailyScoreCalculationTask started", slog.String("time", time.Now().Format(time.RFC3339)))

	timeout := s.cfg.JobContextTimeout
	if timeout <= 0 {
		log.Warn("scheduler JobContextTimeout not configured; defaulting to 5m")
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	if sErr := s.leaderboardStatSvc.GetDailyContributorScores(ctx); sErr != nil {
		log.Error("failed to run dailyScoreCalculationTask", slog.String("error", sErr.Error()))
		return
	}

	log.Info("dailyScoreCalculationTask completed successfully")
}

func (s *Scheduler) publicLeaderboardJob(parentCtx context.Context) error {
	log := logger.L()

	publicLeaderboardJob, err := s.sch.NewJob(
		gocron.CronJob(s.cfg.PublicLeaderboardCron, false),
		gocron.NewTask(func() { s.publicLeaderboardTask(parentCtx) }),
		gocron.WithSingletonMode(gocron.LimitModeWait),
		gocron.WithName("public-leaderboard-update"),
		gocron.WithTags("leaderboardstat-service"),
	)
	if err != nil {
		return fmt.Errorf("failed to create public leaderboard job: %w", err)
	}

	log.Info("publicLeaderboard job created",
		slog.String("name", publicLeaderboardJob.Name()),
		slog.String("uuid", publicLeaderboardJob.ID().String()),
		slog.Any("tags", publicLeaderboardJob.Tags()),
		slog.String("crontab", s.cfg.PublicLeaderboardCron),
	)

	return nil
}

func (s *Scheduler) publicLeaderboardTask(parentCtx context.Context) {
	log := logger.L()

	log.Info("publicLeaderboardTask started", slog.String("time", time.Now().Format(time.RFC3339)))

	ctx, cancel := context.WithTimeout(parentCtx, s.cfg.JobContextTimeout)
	defer cancel()

	if sErr := s.leaderboardStatSvc.SetPublicLeaderboard(ctx); sErr != nil {
		log.Error("failed to run publicLeaderboardTask", slog.String("error", sErr.Error()))
		return
	}

	log.Info("publicLeaderboardTask completed successfully")
}

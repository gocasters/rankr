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

// Job names, as shown in logs. See docs/leaderboardstat-crons.md.
const (
	// JobNamePublicLeaderboard rebuilds the per-project public leaderboard
	// in Redis that GET /v1/leaderboard/public/:project_id serves.
	JobNamePublicLeaderboard = "public-leaderboard-update"
	// JobNameDailyScoreCalculation copies today's ("daily" timeframe) scores
	// from leaderboardscoring into Postgres. "Daily" is the data window,
	// not how often the job runs.
	JobNameDailyScoreCalculation = "daily-score-calculation"

	jobTag = "leaderboardstat-service"
)

// Defaults used when a key is missing from scheduler_cfg. The shipped
// deploy/leaderboardstat/*/config.yml files override both schedules.
const (
	DefaultPublicLeaderboardCron     = "*/3 * * * *" // every 3 minutes
	DefaultDailyScoreCalculationCron = "0 2 * * *"   // daily at 02:00
	DefaultJobContextTimeout         = 3 * time.Minute
)

// Config is the scheduler_cfg block of the leaderboardstat config.
// Crontabs have five fields (minute hour day month weekday) and run in the
// process's local time zone. Each key can be overridden by an environment
// variable, e.g. STAT_SCHEDULER_CFG__PUBLIC_LEADERBOARD_CRON.
type Config struct {
	// DailyScoreCalculationCron schedules JobNameDailyScoreCalculation.
	DailyScoreCalculationCron string `koanf:"daily_score_calculation_cron"`
	// PublicLeaderboardCron schedules JobNamePublicLeaderboard.
	PublicLeaderboardCron string `koanf:"public_leaderboard_cron"`
	// JobContextTimeout bounds a single run of either job.
	JobContextTimeout time.Duration `koanf:"job_context_timeout"`
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
		schedulerCfg.PublicLeaderboardCron = DefaultPublicLeaderboardCron
	}
	if schedulerCfg.JobContextTimeout <= 0 {
		schedulerCfg.JobContextTimeout = DefaultJobContextTimeout
	}
	if schedulerCfg.DailyScoreCalculationCron == "" {
		schedulerCfg.DailyScoreCalculationCron = DefaultDailyScoreCalculationCron
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
		s.cfg.DailyScoreCalculationCron = DefaultDailyScoreCalculationCron
	}

	dailyJob, err := s.sch.NewJob(
		gocron.CronJob(s.cfg.DailyScoreCalculationCron, false),
		gocron.NewTask(func() { s.dailyScoreCalculationTask(parentCtx) }),
		gocron.WithSingletonMode(gocron.LimitModeWait),
		gocron.WithName(JobNameDailyScoreCalculation),
		gocron.WithTags(jobTag),
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
		gocron.WithName(JobNamePublicLeaderboard),
		gocron.WithTags(jobTag),
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

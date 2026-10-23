package scheduler

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/gocasters/rankr/adapter/project"
	"github.com/gocasters/rankr/pkg/logger"
	"log/slog"
	"sync"
	"time"
)

type ProjectService interface {
	ListProjects(ctx context.Context, req *project.ListProjectsRequest) (*project.ListProjectsResponse, error)
}

type LeaderboardService interface {
	LeaderboardSnapshot(ctx context.Context, projectIDs []string) error
}

type Config struct {
	SnapshotCrontab           string        `koanf:"snapshot_crontab"`
	SnapshotJobContextTimeout time.Duration `koanf:"snapshot_job_context_timeout"`
	PageSizeListProject       int32         `koanf:"page_size_list_project"`
}

type Scheduler struct {
	sch            gocron.Scheduler
	leaderboardSvc LeaderboardService
	cfg            Config
	project        ProjectService
}

func New(leaderboardSvc LeaderboardService, projectClient ProjectService, schedulerCfg Config) Scheduler {

	sch, err := gocron.NewScheduler(gocron.WithLocation(time.Local))
	if err != nil {
		logger.L().Error("failed to create Scheduler", slog.String("error", err.Error()))
		panic(err)
	}

	return Scheduler{
		sch:            sch,
		leaderboardSvc: leaderboardSvc,
		project:        projectClient,
		cfg:            schedulerCfg,
	}
}

func (s *Scheduler) Start(ctx context.Context, wg *sync.WaitGroup) {
	log := logger.L()
	defer wg.Done()

	log.Info("scheduler started")

	if err := s.leaderboardSnapshotJob(ctx); err != nil {
		log.Error("failed to create snapshot job", slog.String("error", err.Error()))
	}

	s.sch.Start()

	<-ctx.Done()

	log.Info("scheduler exiting...")

	if sErr := s.sch.Shutdown(); sErr != nil {
		log.Warn("cannot successfully shutdown scheduler", slog.String("error", sErr.Error()))
	}
}

func (s *Scheduler) leaderboardSnapshotJob(parentCtx context.Context) error {
	log := logger.L()

	if s.cfg.SnapshotCrontab == "" {
		return fmt.Errorf("snapshot_crontab is empty")
	}

	snapshotJob, err := s.sch.NewJob(
		gocron.CronJob(s.cfg.SnapshotCrontab, false),
		gocron.NewTask(func() { s.snapshotLeaderboardTask(parentCtx) }),
		gocron.WithSingletonMode(gocron.LimitModeWait),
		gocron.WithName("snapshot-leaderboard"),
		gocron.WithTags("leaderboardscoring-service"),
	)
	if err != nil {
		return fmt.Errorf("failed to create snapshot job: %w", err)
	}

	log.Info("leaderboardSnapshot job created",
		slog.String("name", snapshotJob.Name()),
		slog.String("uuid", snapshotJob.ID().String()),
		slog.Any("tags", snapshotJob.Tags()),
	)

	return nil
}

func (s *Scheduler) snapshotLeaderboardTask(parentCtx context.Context) {
	log := logger.L()

	log.Debug(fmt.Sprintf(
		"snapshotLeaderboardTask started at: %s",
		time.Now().Format(time.RFC3339),
	))

	ctx, cancel := context.WithTimeout(parentCtx, s.cfg.SnapshotJobContextTimeout)
	defer cancel()

	var offset int32
	projectIDs := make([]string, 0)

	for {
		pRes, err := s.project.ListProjects(ctx, &project.ListProjectsRequest{
			PageSize: s.cfg.PageSizeListProject,
			Offset:   offset,
			Status:   project.StatusActive,
		})
		if err != nil {
			log.Warn(
				"failed to get list active projects",
				slog.String("error", err.Error()),
			)
			return
		}

		if len(pRes.Projects) == 0 {
			// No projects mean pagination is complete unless TotalCount indicates missing data
			if offset < pRes.TotalCount {
				log.Warn(
					"project pagination made no progress",
					slog.Int("offset", int(offset)),
					slog.Int("total_count", int(pRes.TotalCount)),
				)
				return
			}

			break
		}

		for _, p := range pRes.Projects {
			projectIDs = append(projectIDs, p.ProjectID)
		}

		offset += int32(len(pRes.Projects))

		if offset >= pRes.TotalCount {
			break
		}
	}

	if sErr := s.leaderboardSvc.LeaderboardSnapshot(ctx, projectIDs); sErr != nil {
		log.Warn(
			"can not successfully run snapshotLeaderboardTask",
			slog.String("error", sErr.Error()),
		)
		return
	}

	log.Debug("snapshotLeaderboardTask completed successfully")
}

package userprofile

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
)

type ContributorRPC interface {
	GetProfileInfo(ctx context.Context, contributorID int64) (ContributorInfo, error)
}

type TaskRPC interface {
	GetTasks(ctx context.Context, contributorID int64) ([]Task, error)
}

type LeaderboardStatRPC interface {
	GetContributorStat(ctx context.Context, contributorID int64) (ContributorStat, error)
}

type Config struct {
	UpstreamTimeout time.Duration `koanf:"upstream_time_out"`
}

type Service struct {
	contributor     ContributorRPC
	task            TaskRPC
	leaderboardStat LeaderboardStatRPC
	validator       Validator
	config          Config
}

func NewService(
	contributorInfo ContributorRPC,
	task TaskRPC,
	leaderboardStat LeaderboardStatRPC,
	validator Validator,
	config Config,
) Service {
	return Service{
		contributor:     contributorInfo,
		task:            task,
		leaderboardStat: leaderboardStat,
		validator:       validator,
		config:          config,
	}
}

func (s Service) ContributorProfile(ctx context.Context, contributorID int64) (*ProfileResponse, error) {

	log := logger.L()
	log.Info("Starting contributor profile", slog.Int64("contributor_id", contributorID))

	var contributorInfo ContributorInfo
	var tasks = make([]Task, 0)
	var contributorStat ContributorStat

	var cErr, tErr, lErr error

	sCtx, cancel := context.WithTimeout(ctx, s.config.UpstreamTimeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		ci, err := s.contributor.GetProfileInfo(sCtx, contributorID)
		if err != nil {
			cErr = err
			return
		}

		contributorInfo = ci
	}()

	go func() {
		defer wg.Done()
		cs, err := s.leaderboardStat.GetContributorStat(sCtx, contributorID)
		if err != nil {
			lErr = err
			return
		}

		contributorStat = cs
	}()

	go func() {
		defer wg.Done()
		ts, err := s.task.GetTasks(sCtx, contributorID)
		if err != nil {
			tErr = err
			return
		}

		tasks = ts

	}()

	wg.Wait()

	if cErr != nil {
		if errors.Is(cErr, context.DeadlineExceeded) {
			log.Error("time out contributor service", slog.Any("error", cErr))

		} else {
			log.Error("failed to get contributor info", slog.Any("error", cErr))
		}
	}

	if lErr != nil {
		if errors.Is(lErr, context.DeadlineExceeded) {
			log.Error("time out leaderboard stat service", slog.Any("error", lErr))

		} else {
			log.Error("failed to get contributor leaderboard stat", slog.Any("error", lErr))
		}
	}

	if tErr != nil {
		if errors.Is(tErr, context.DeadlineExceeded) {
			log.Error("time out task service", slog.Any("error", tErr))

		} else {
			log.Error("failed to get contributor tasks", slog.Any("error", tErr))
		}
	}

	return &ProfileResponse{
		Profile: Profile{
			ContributorInfo: contributorInfo,
			Tasks:           tasks,
			ContributorStat: contributorStat,
		},
	}, nil
}

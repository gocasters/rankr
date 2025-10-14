package userprofile

import (
	"context"
	"github.com/gocasters/rankr/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type ContributorRPC interface {
	GetProfileInfo(ctx context.Context, userID int64) (ContributorInfo, error)
}

type TaskRPC interface {
	GetTasks(ctx context.Context, userID int64) ([]Task, error)
}

type LeaderboardStatRPC interface {
	GetContributorStat(ctx context.Context, userID int64) (ContributorStat, error)
}

type Service struct {
	contributorInfo ContributorRPC
	task            TaskRPC
	leaderboardStat LeaderboardStatRPC
	validator       Validator
}

func NewService(
	contributorInfo ContributorRPC,
	task TaskRPC,
	leaderboardStat LeaderboardStatRPC,
	validator Validator,
) Service {
	return Service{
		contributorInfo: contributorInfo,
		task:            task,
		leaderboardStat: leaderboardStat,
		validator:       validator,
	}
}

func (s Service) ContributorProfile(ctx context.Context, contributorID int64) (*ProfileResponse, error) {

	var contributorInfo ContributorInfo
	var tasks = make([]Task, 0)
	var contributorStat ContributorStat

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		ci, err := s.contributorInfo.GetProfileInfo(gctx, contributorID)
		if err != nil {
			return err
		}

		contributorInfo = ci

		return nil
	})

	g.Go(func() error {
		ts, err := s.task.GetTasks(gctx, contributorID)
		if err != nil {
			return err
		}

		tasks = ts

		return nil
	})

	g.Go(func() error {
		cs, err := s.leaderboardStat.GetContributorStat(gctx, contributorID)
		if err != nil {
			return err
		}

		contributorStat = cs

		return nil
	})

	if err := g.Wait(); err != nil {
		logger.L().Error("userprofile-get-profile", "error", err)

		return nil, err
	}

	return &ProfileResponse{
		Profile: Profile{
			ContributorInfo: contributorInfo,
			Tasks:           tasks,
			ContributorStat: contributorStat,
		},
	}, nil
}

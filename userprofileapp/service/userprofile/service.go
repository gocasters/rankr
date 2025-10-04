package userprofile

import (
	"context"
	"github.com/gocasters/rankr/pkg/logger"
)

type RPCRepository interface {
	ContributorRPC
	TaskRPC
	LeaderboardStatRPC
}

type ContributorRPC interface {
	GetProfileInfo(ctx context.Context, userID int64) (ContributorInfo, error)
}

type TaskRPC interface {
	GetTasks(ctx context.Context, userID int64) ([]Task, error)
}

type LeaderboardStatRPC interface {
	GetUserStat(ctx context.Context, userID int64) (ContributorStat, error)
}

type Service struct {
	rpcRepo   RPCRepository
	validator Validator
}

func NewService(rpcRepo RPCRepository, validator Validator) Service {
	return Service{
		rpcRepo:   rpcRepo,
		validator: validator,
	}
}

func (s Service) GetUserProfile(ctx context.Context, contributorID int64) (*ProfileResponse, error) {

	contributorInfo, err := s.rpcRepo.GetProfileInfo(ctx, contributorID)
	if err != nil {
		logger.L().Error("userprofile-get-user-profile", "error: ", err)
		return nil, err
	}

	tasks, err := s.rpcRepo.GetTasks(ctx, contributorID)
	if err != nil {
		logger.L().Error("userprofile-get-user-profile", "error: ", err)
		return nil, err
	}

	contributorStat, err := s.rpcRepo.GetUserStat(ctx, contributorID)
	if err != nil {
		logger.L().Error("userprofile-get-user-profile", "error: ", err)
		return nil, err
	}

	userProfile := ProfileResponse{
		Profile: Profile{
			ContributorInfo: contributorInfo,
			Tasks:           tasks,
			ContributorStat: contributorStat,
		},
	}

	return &userProfile, nil
}

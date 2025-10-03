package userprofile

import (
	"context"
	"github.com/gocasters/rankr/pkg/logger"
)

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
	contributorRPC ContributorRPC
	taskPRC        TaskRPC
	leaderboardRPC LeaderboardStatRPC
	validator      Validator
}

func NewService(contributorRPC ContributorRPC, taskRPC TaskRPC, leaderboardRPC LeaderboardStatRPC, validator Validator) Service {
	return Service{
		contributorRPC: contributorRPC,
		taskPRC:        taskRPC,
		leaderboardRPC: leaderboardRPC,
		validator:      validator,
	}
}

func (s Service) GetUserProfile(ctx context.Context, contributorID int64) (*ProfileResponse, error) {

	contributorInfo, err := s.contributorRPC.GetProfileInfo(ctx, contributorID)
	if err != nil {
		logger.L().Error("userprofile-get-user-profile", "error: ", err)
		return nil, err
	}

	tasks, err := s.taskPRC.GetTasks(ctx, contributorID)
	if err != nil {
		logger.L().Error("userprofile-get-user-profile", "error: ", err)
		return nil, err
	}

	contributorStat, err := s.leaderboardRPC.GetUserStat(ctx, contributorID)
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

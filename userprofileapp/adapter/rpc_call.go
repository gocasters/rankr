package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type RPC struct {
	contributorInfoRPC ContributorInfoRPC
	tasksRPC           TaskRPC
	contributorStatRPC ContributorStatRPC
}

func NewRPC(
	contributorInfoRPC ContributorInfoRPC,
	contributorStatRPC ContributorStatRPC,
	taskRPC TaskRPC,
) RPC {
	return RPC{
		contributorInfoRPC: contributorInfoRPC,
		tasksRPC:           taskRPC,
		contributorStatRPC: contributorStatRPC,
	}
}

func (a RPC) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	return a.tasksRPC.getTasks(ctx, userID)
}

func (a RPC) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	return a.contributorInfoRPC.getProfileInfo(ctx, userID)
}

func (a RPC) GetContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	return a.contributorStatRPC.getContributorStat(ctx, userID)
}

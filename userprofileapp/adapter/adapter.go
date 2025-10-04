package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type RPCAdapter struct {
	contributorInfoRPC contributorInfoRPC
	tasksRPC           tasksRPC
	contributorStatRPC contributorStatRPC
}

func NewRPCAdapter() RPCAdapter {
	return RPCAdapter{
		contributorInfoRPC: newContributorInfoRPC(),
		tasksRPC:           newTasksRPC(),
		contributorStatRPC: newContributorStatRPC(),
	}
}

func (a RPCAdapter) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	return a.tasksRPC.getTasks(ctx, userID)
}

func (a RPCAdapter) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	return a.contributorInfoRPC.getProfileInfo(ctx, userID)
}

func (a RPCAdapter) GetContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	return a.contributorStatRPC.getContributorStat(ctx, userID)
}

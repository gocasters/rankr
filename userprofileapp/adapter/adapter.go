package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type RPCAdapter struct {
	contributorRPC     ContributorRPC
	tasksRPC           TasksRPC
	contributorStatRPC ContributorStatRPC
}

func NewRPCAdapter() RPCAdapter {
	return RPCAdapter{
		contributorRPC:     NewContributorRPC(),
		tasksRPC:           NewTasksRPC(),
		contributorStatRPC: NewContributorStatRPC(),
	}
}

func (a RPCAdapter) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// TODO: Implement me

	return nil, fmt.Errorf("implement me. userprofileapp/adapter/task_rpc.go GetTasks")
}

func (a RPCAdapter) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	// TODO: Implement me

	return userprofile.ContributorInfo{}, fmt.Errorf("implement me. userprofileapp/adapter/contributor_rpc.go GetProfileInfo")
}

func (a RPCAdapter) GetContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	// TODO: Implement me

	return userprofile.ContributorStat{}, fmt.Errorf("implement me. userprofileapp/adapter/contributor_rpc.go GetContributorStat")
}

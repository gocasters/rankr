package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type tasksRPC struct{}

func newTasksRPC() tasksRPC {
	return tasksRPC{}
}

func (a tasksRPC) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// TODO: Implement me

	return nil, fmt.Errorf("implement me. userprofileapp/adapter/task_rpc.go GetTasks")
}

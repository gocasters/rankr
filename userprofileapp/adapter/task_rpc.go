package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type TasksRPC struct{}

func NewTasksRPC() TasksRPC {
	return TasksRPC{}
}

func (a TasksRPC) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// TODO: Implement me

	return nil, fmt.Errorf("implement me. userprofileapp/adapter/task_rpc.go GetTasks")
}

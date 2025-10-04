package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type tasksRPC struct{}

func newTasksRPC() tasksRPC {
	return tasksRPC{}
}

// GetTasks fetches contributor's tasks
// TODO: Implement actual RPC call to taskapp
func (a tasksRPC) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// TODO: Implement me

	return []userprofile.Task{}, nil
}

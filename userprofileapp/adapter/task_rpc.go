package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type tasksRPC struct{}

func newTasksRPC() tasksRPC {
	return tasksRPC{}
}

// getTasks fetches contributor's tasks
// TODO: Implement actual RPC call to taskapp
func (a tasksRPC) getTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// Placeholder implementation

	return []userprofile.Task{}, nil
}

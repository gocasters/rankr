package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type TaskRPC struct{}

func NewTaskRPC() TaskRPC {
	return TaskRPC{}
}

// GetTasks fetches contributor's tasks
// TODO: Implement actual RPC call to task app
func (a TaskRPC) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// Placeholder implementation

	return []userprofile.Task{}, nil
}

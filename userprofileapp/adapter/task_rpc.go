package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type TaskRPC struct{}

func NewTaskRPC() TaskRPC {
	return TaskRPC{}
}

// getTasks fetches contributor's tasks
// TODO: Implement actual RPC call to taskapp
func (a TaskRPC) getTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	// Placeholder implementation

	return []userprofile.Task{}, nil
}

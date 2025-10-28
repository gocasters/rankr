package task

import (
	"time"
)

type CreateTaskParam struct {
	GithubID       int64
	IssueNumber    int
	Title          string
	State          string
	RepositoryName string
	Labels         []string
	CreatedAt      time.Time
}

type UpdateTaskParam struct {
	IssueNumber    int
	RepositoryName string
	State          string
	ClosedAt       time.Time
}
